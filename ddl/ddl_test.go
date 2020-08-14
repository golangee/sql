package ddl

import (
	"fmt"
	"github.com/golangee/reflectplus/src"
	"testing"
)

func TestAppSpecification(t *testing.T) {

	NewRepositories("migrations").
		Comment("...contains all migration related repositories and their available implementations at the infrastructure level").
		Migrate(
			// TODO sql needs 10 tables but NoSQL needs 1 table: how to express that?
			NewTable("migration_schema",
				NewColumn("group").Varchar(255).NotNull().Comment("Unique domain or bounded context name"),
				NewColumn("version").Int64().NotNull(),
				NewColumn("script").Varchar(255).Nullable(),
				NewColumn("type").Varchar(12).NotNull(),
				NewColumn("checksum").Char(64).NotNull(),
				NewColumn("applied_at").Timestamp().NotNull(),
				NewColumn("execution_duration").Duration().NotNull(),
				NewColumn("status").Enum("success", "failed", "pending", "executing").NotNull(),
			).PrimaryKey("group", "version").Comment("Contains all database migrations"),
		).
		Add(
			NewRepository("stuff_repo").Comment("...returns things about stuff.").
				AddQueries(
					InsertOne("migration_schema"),
					ReadOne("migration_schema"),
					ReadAll("migration_schema"),
					UpdateOne("migration_schema"),
					DeleteOne("migration_schema"),
					DeleteAll("migration_schema"),
					CountAll("migration_schema"),

					NewQuery("find_some_stuff",
						NewRawQuery(MySQL, "SELECT `script, applied FROM migration_schema WHERE status = :name"),
						NewRawQuery(PostgreSQL, "SELECT script, applied FROM migration_schema WHERE status = :name"),
					).Comment("... finds some cool stuff").
						Input("name", Text).
						Input("other", Int64).
						OutputMany(Text, Bool),
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

	//TODO we need a check to deduplicate enums and detect definition clashes
	goStruct, enums := table.golangRow()

	srcFile := src.NewFile("blub").AddTypes(
		goStruct,
		enums[0],
		dbCTXInterface(),
	)
	fmt.Println(srcFile.String())
}
