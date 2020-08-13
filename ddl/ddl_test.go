package ddl

import (
	"fmt"
	"github.com/golangee/reflectplus/src"
	"testing"
)

func TestAppSpecification(t *testing.T) {

	table := NewTable("migration_schema",
		NewColumn("group").Varchar(255).NotNull().Comment("Unique domain or bounded context name"),
		NewColumn("version").Int64().NotNull(),
		NewColumn("script").Varchar(255).NotNull(),
		NewColumn("type").Varchar(12).NotNull(),
		NewColumn("checksum").Char(64).NotNull(),
		NewColumn("applied_at").Timestamp().NotNull(),
		NewColumn("execution_duration").Duration().NotNull(),
		NewColumn("status").Enum("success", "failed", "pending", "executing").NotNull(),
	).PrimaryKey("group", "version").Comment("Contains all database migrations")

	NewBoundedContext("migrations", nil).
		SetDoc("...contains all those datatypes and the repository for the migrations.").
		AddQueries(
			NewRawQuery(DialectMySQL, "find_stuff_by_status", "SELECT group, version FROM migration_schema WHERE status = :status").
				StructParams(
					src.NewTypeDecl("string"),
				).
				MapResult(
					src.NewTypeDecl("string"),
					src.NewTypeDecl("int64"),
				),

			NewRawQuery(DialectMySQL, "find_by_status", "SELECT group, version FROM migration_schema WHERE status = :status").
				Params(
					src.NewTypeDecl("string"),
				).
				StructResult(
					src.NewTypeDecl("string"),
					src.NewTypeDecl("int64"),
				),

			NewRawQuery(DialectMySQL, "delete", "DELETE FROM migration_schema WHERE group=:group AND version =:version").
				Params(
					src.NewTypeDecl("string"),
					src.NewTypeDecl("string"),
				),

			NewRawQuery(DialectMySQL, "search", "SELECT id FROM migration_schema").
				Results(
					src.NewTypeDecl("int32"),
				),
		)

	s, err := ToString(table.AsMySQL)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(s)

	/*
		s, err = ToString(table.AsPlantUML)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(s)*/

	goStruct, enums := table.golangRow()

	srcFile := src.NewFile("blub").AddTypes(
		goStruct,
		enums[0],
		dbCTXInterface(),
	)
	fmt.Println(srcFile.String())
}
