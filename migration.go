package sql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/golangee/reflectplus"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const createMigrationTable = `CREATE TABLE IF NOT EXISTS "migration_schema_history"
(
    "group"              VARCHAR(255) NOT NULL,
    "version"            BIGINT       NOT NULL,
    "script"             VARCHAR(255) NOT NULL,
    "type"               VARCHAR(12)  NOT NULL,
    "checksum"           CHAR(64)     NOT NULL,
    "applied_at"         TIMESTAMP    NOT NULL,
    "execution_duration" BIGINT       NOT NULL,
    "status"             VARCHAR(12)  NOT NULL,
    "log"                TEXT         NOT NULL,
    PRIMARY KEY ("group", "version")
)`

// MustMigrate panics, if the migrations cannot be applied.
// Creates a transaction and tries a rollback, before bailing out. Delegates to #Migrate() and auto detects dialect.
func MustMigrate(db *sql.DB) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	dialect, err := DetectDialect(tx)
	if err != nil {
		panic(err)
	}
	if err := Migrate(dialect, tx); err != nil {
		if suppressedErr := tx.Rollback(); suppressedErr != nil {
			fmt.Println(suppressedErr.Error())
		}
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

// Migrate grabs all @ee.sql.Schema() annotations from all available repositories
// and tries to apply them. Each schema statement is verified to proof that the
// migration process is repeatable.
func Migrate(dialect Dialect, dbtx DBTX) error {
	var res []Migration
	for _, iface := range reflectplus.Interfaces() {

		if iface.GetAnnotations().Has(AnnotationRepository) {
			stereotype, err := iface.GetAnnotations().MustOne(AnnotationRepository)
			if err != nil {
				return err
			}
			groupName := stereotype.Value()
			migrations := filterByDialect(dialect, iface.GetAnnotations().FindAll(AnnotationSchema))
			if len(migrations) > 0 && groupName == "" {
				return reflectplus.PositionalError(iface, fmt.Errorf("%s must provide a default name if used with %s", AnnotationRepository, AnnotationSchema))
			}

			for idx, schema := range migrations {
				stmt := strings.TrimSpace(schema.Value())
				if stmt == "" {
					return reflectplus.PositionalError(schema, fmt.Errorf("value of '%s' contains the sql statement and must not be empty", AnnotationSchema))
				}

				group := schema.OptString(GroupValue, groupName)
				version := schema.OptInt(VersionValue, idx)

				migration := Migration{
					Group:      group,
					Version:    int64(version),
					Statements: []string{stmt},
					ScriptName: iface.Pos.Filename + ":" + strconv.Itoa(iface.Pos.Line),
				}
				res = append(res, migration)

			}
		}
	}
	return ApplyMigrations(dialect, dbtx, res...)
}

type MigrationType string
type MigrationStatus string

// actually a bad code smell but this is intentional, bcause we explicitly don't want concurrent migrations
// within a single process (actually not even between processes), for your own brains sake.
var mutex sync.Mutex

const (
	SQL       MigrationType   = "sql"
	Success   MigrationStatus = "success"
	Failed    MigrationStatus = "failed"
	Pending   MigrationStatus = "pending"
	Executing MigrationStatus = "executing"
)

type MigrationStatusEntry struct {
	Group             string
	Version           int64
	Script            string
	Type              MigrationType
	Checksum          string
	AppliedAt         time.Time
	ExecutionDuration time.Duration
	Status            MigrationStatus
	Log               string
}

type Migration = struct {
	Group      string
	Version    int64
	Statements []string
	ScriptName string
}

func hash(m Migration) string {
	sum := sha256.Sum256([]byte(strings.Join(m.Statements, ";")))
	return hex.EncodeToString(sum[:])
}

func createHistoryTable(dialect Dialect, db DBTX) error {
	// this schema should be "cross platform"
	if _, err := db.ExecContext(context.Background(), createMigrationTable); err != nil {
		fmt.Println(createMigrationTable)
		return fmt.Errorf("cannot create migration table: %w", err)
	}
	return nil
}

// ApplyMigrations calculates which migrations needs to be applied and tries to apply the missing ones.
func ApplyMigrations(dialect Dialect, db DBTX, migrations ...Migration) error {
	mutex.Lock()
	defer mutex.Unlock()

	err := createHistoryTable(dialect, db)
	if err != nil {
		return err
	}

	entries, err := SchemaHistory(db)
	if err != nil {
		return fmt.Errorf("cannot get history: %w", err)
	}

	for _, entry := range entries {
		if entry.Status != Success {
			return fmt.Errorf("migrations are dirty. Needs manual fix: %+v", entry)
		}
	}

	// group by and pick those things, which have not been applied yet
	groups := make(map[string][]Migration)
	for _, migration := range migrations {
		alreadyApplied := false
		for _, entry := range entries {
			if migration.Group == entry.Group {
				if migration.Version == entry.Version {
					if hash(migration) != entry.Checksum {
						return fmt.Errorf("an already applied migration (%s) has been modified. Needs manual fix: %v vs %v", migration.ScriptName, entry, migration)
					}
					//fmt.Println(migration.Group, hash(migration), "=>", entry.Checksum,migration.Statements)
					alreadyApplied = true
					//fmt.Printf("migration already applied: %s.%d\n", migration.Group, migration.Version)
					break
				}
			}
		}
		if !alreadyApplied {
			candidatesPerGroup := groups[migration.Group]
			candidatesPerGroup = append(candidatesPerGroup, migration)
			groups[migration.Group] = candidatesPerGroup
		}
	}

	// uniqueness check
	for _, candidates := range groups {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Version < candidates[j].Version
		})
		// ensure unique constraints per group
		strictMonotonicVersion := int64(-1)
		for _, m := range candidates {
			if m.Version <= strictMonotonicVersion {
				return fmt.Errorf("the version must be >=0 and unique: %v", m)
			}
		}
	}

	// actually apply the missing migrations
	for _, candidates := range groups {
		for _, migration := range candidates {
			entry := MigrationStatusEntry{
				Group:             migration.Group,
				Version:           migration.Version,
				Script:            migration.ScriptName,
				Type:              SQL,
				Checksum:          hash(migration),
				AppliedAt:         time.Now(),
				ExecutionDuration: 0,
				Status:            Executing,
			}

			start := time.Now()
			if err := insert(dialect, db, entry); err != nil {
				return fmt.Errorf("failed to insert history entry: %w", err)
			}

			if err := execute(db, migration); err != nil {
				entry.Log = err.Error()
				entry.Status = Failed
				_ = update(dialect, db, entry)
				return fmt.Errorf("failed to execute migration %s.%d: %w", migration.Group, migration.Version, err)
			}

			entry.Status = Success
			entry.ExecutionDuration = time.Now().Sub(start)

			if err := update(dialect, db, entry); err != nil {
				return fmt.Errorf("failed to update history migration: %w", err)
			}
		}
	}
	return nil
}

func insert(dialect Dialect, tx DBTX, entry MigrationStatusEntry) error {
	stmt, err := NamedParameterStatement(`INSERT INTO "migration_schema_history" ("group", "version", "script", "type", "checksum", "applied_at", "execution_duration", "status", "log") VALUES (:1,:2,:3,:4,:5,:6,:7,:8,:9)`).
		Prepare(dialect, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"})

	if err != nil {
		return err
	}

	if _, err := stmt.ExecContext(tx, context.Background(), entry.Group, entry.Version, entry.Script, entry.Type, entry.Checksum, entry.AppliedAt, entry.ExecutionDuration, entry.Status, entry.Log); err != nil {
		return err
	}
	return nil
}

func update(dialect Dialect, tx DBTX, entry MigrationStatusEntry) error {
	stmt, err := NamedParameterStatement(`UPDATE "migration_schema_history" SET "script"=:1, "type"=:2, "checksum"=:3, "applied_at"=:4, "execution_duration"=:5, "status"=:6, "log"=:7 WHERE "group"=:8 and "version"=:9`).
		Prepare(dialect, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"})

	if err != nil {
		return err
	}

	if _, err := stmt.ExecContext(tx, context.Background(), entry.Script, entry.Type, entry.Checksum, entry.AppliedAt, entry.ExecutionDuration, entry.Status, entry.Log, entry.Group, entry.Version); err != nil {
		return err
	}
	return nil
}

func execute(tx DBTX, migration Migration) error {
	for _, stmt := range migration.Statements {
		if _, err := tx.ExecContext(context.Background(), stmt); err != nil {
			return fmt.Errorf("failed to execute statement '%s': %w", stmt, err)
		}
	}
	return nil
}

// SchemaStatus returns all applied migration or schema scripts and their according states.
func SchemaHistory(tx DBTX) ([]MigrationStatusEntry, error) {
	rows, err := tx.QueryContext(context.Background(), "SELECT `group`, `version`, `script`, `type`, `checksum`, `applied_at`, `execution_duration`, `status`,`log` FROM `migration_schema_history`")
	if err != nil {
		return nil, fmt.Errorf("cannot select history: %w", err)
	}
	defer rows.Close()

	var res []MigrationStatusEntry
	for rows.Next() {
		entry := MigrationStatusEntry{}
		err = rows.Scan(&entry.Group, &entry.Version, &entry.Script, &entry.Type, &entry.Checksum, &entry.AppliedAt, &entry.ExecutionDuration, &entry.Status, &entry.Log)
		if err != nil {
			return res, fmt.Errorf("cannot scan entry: %w", err)
		}
		res = append(res, entry)
	}
	if rows.Err() != nil {
		return res, fmt.Errorf("cannot scan history: %w", err)
	}
	return res, nil
}
