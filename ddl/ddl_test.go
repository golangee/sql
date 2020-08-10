package ddl

import (
	"fmt"
	"github.com/golangee/reflectplus/src"
	"testing"
)

func TestChange(t *testing.T) {
	table := NewTable("migration_schema",
		NewColumn("group", String).Len(255).NotNull().Doc("Unique domain or bounded context name"),
		NewColumn("version", Int64).NotNull(),
		NewColumn("script", String).Len(255).NotNull(),
		NewColumn("checksum", String).Len(12).NotNull(),
		NewColumn("applied_at", Timestamp).NotNull(),
		NewColumn("execution_duration", Int64).NotNull(),
		NewColumn("status", String).Len(12).NotNull(),
	).PrimaryKey("group", "version").Doc("Contains all database migrations")

	s, err := ToString(table.AsMySQL)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(s)

	s, err = ToString(table.AsPlantUML)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(s)

	srcFile := src.NewFile("blub").AddTypes(
		table.AsGoStruct(),
		table.AsGoMySQLCRUDRepository(),
	)
	fmt.Println(srcFile.String())
}
