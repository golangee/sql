// Copyright 2021 The Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mysql_test

import (
	"fmt"
	"github.com/golangee/sql/ddl"
	"github.com/golangee/sql/dialect/mysql"
	"github.com/golangee/sql/internal"
	"io/ioutil"
	"testing"
)

func loadSql(fname string) string {
	sqlBytes, err := ioutil.ReadFile("testdata/" + fname)
	if err != nil {
		panic(err)
	}
	return string(sqlBytes)
}

func TestParseMusic(t *testing.T) {
	sql := loadSql("music.sql")

	expectedTables := expectedMusicTables()

	actualResult, err := mysql.Parse(sql)
	if err != nil {
		t.Fatal(err)
	}

	// Did we get all the tables?
	if len(actualResult.Tables) != len(expectedTables) {
		t.Fatalf("Expected %v tables, but got %v", len(expectedTables), len(actualResult.Tables))
	}

	// Compare outputs
	for i := 0; i < len(expectedTables); i++ {
		actual := actualResult.Tables[i]
		expected := expectedTables[i]
		internal.DiffCompare(t, actual, expected, fmt.Sprintf("table %s", actual.Name))
	}
}

func TestParseAlter(t *testing.T) {
	sql := loadSql("alter-user.sql")

	expectedAlters := expectedUserAlterStatements()

	actualResult, err := mysql.Parse(sql)
	if err != nil {
		t.Fatal(err)
	}

	// Did we get all the statements?
	if len(actualResult.AlterStatements) != len(expectedAlters) {
		t.Fatalf("Expected %v statements, but got %v", len(expectedAlters), len(actualResult.AlterStatements))
	}

	// Compare outputs
	for i := 0; i < len(expectedAlters); i++ {
		actual := actualResult.AlterStatements[i]
		expected := expectedAlters[i]
		internal.DiffCompare(t, actual, expected, fmt.Sprintf("statement #%d", i))
	}
}

// Returns a list of ALTER TABLE statements, that should be present in testdata/alter-user.sql.
func expectedUserAlterStatements() []ddl.AlterStatement {
	return []ddl.AlterStatement{
		ddl.AlterAddColumn{
			Table:  "User",
			Column: ddl.Column{Name: "BirthDate", Type: "DATE"},
		},
		ddl.AlterAddColumn{
			Table:  "User",
			Column: ddl.Column{Name: "Comment", Type: "TEXT"},
		},
		ddl.AlterAddColumn{
			Table:  "User",
			Column: ddl.Column{Name: "BirthYear", Type: "INT"},
		},
		ddl.AlterAddColumn{
			Table:  "User",
			Column: ddl.Column{Name: "Id", Type: "INT", NotNull: true, Default: s("123")},
			First:  true,
		},
		ddl.AlterAddColumn{
			Table:  "User",
			Column: ddl.Column{Name: "Id", Type: "INT", NotNull: true, Default: s("123")},
			After:  s("BirthDate"),
		},
		ddl.AlterDropColumn{
			Table:  "User",
			Column: "BirthDate",
		},
		ddl.AlterDropColumn{
			Table:  "User",
			Column: "Id",
		},
		ddl.AlterAddIndex{
			Table:  "User",
			Name:   "IndexId",
			Column: "Id",
			Unique: true,
		},
		ddl.AlterDropIndex{
			Table: "User",
			Index: "IndexId",
		},
		ddl.AlterDropIndex{
			Table: "User",
			Index: "IndexId2",
		},
	}
}

// expectedMusicTables returns the expected model from the testdata/music.sql example.
func expectedMusicTables() []ddl.Table {
	return []ddl.Table{
		{
			Name:        "Artist",
			IfNotExists: true,
			Columns: []ddl.Column{
				{Name: "Id", Type: "INT", PrimaryKey: true},
				{Name: "Name", Type: "VARCHAR(255)", NotNull: true, Unique: true},
				{Name: "BirthYear", Type: "INT", NotNull: true},
			},
		},
		{
			Name: "Song",
			Columns: []ddl.Column{
				{Name: "Id", Type: "INT", PrimaryKey: true},
				{Name: "Name", Type: "VARCHAR(255)", NotNull: true},
				{Name: "Album", Type: "INT"},
			},
			ForeignKeys: []ddl.ForeignKeyConstraint{
				{Column: "Album", ReferenceTable: "Album", ReferenceColumn: "Id"},
			},
		},
		{
			Name: "WorkedOn",
			Columns: []ddl.Column{
				{Name: "Artist", Type: "INT", NotNull: true},
				{Name: "Song", Type: "INT", NotNull: true},
			},
			ForeignKeys: []ddl.ForeignKeyConstraint{
				{Name: s("Wrote"), Column: "Artist", ReferenceTable: "Artist", ReferenceColumn: "Id"},
				{Name: s("WrittenBy"), Column: "Song", ReferenceTable: "Song", ReferenceColumn: "Id"},
			},
		},
		{
			Name: "Album",
			Columns: []ddl.Column{
				{Name: "Id", Type: "INT", PrimaryKey: true},
				{Name: "Name", Type: "VARCHAR(255)"},
				{Name: "Year", Type: "INT", Default: s("2000")},
			},
		},
		{
			Name: "Publisher",
			Columns: []ddl.Column{
				{Name: "Id", Type: "INT", PrimaryKey: true},
				{Name: "Uuid", Type: "INT"},
				{Name: "Year", Type: "INT"},
			},
			Keys: []ddl.Key{
				{Name: s("k_uuid"), OnColumn: "Uuid"},
				{OnColumn: "Year"},
			},
		},
	}
}

// Turn a string into a *string. Needed for nullable strings.
func s(s string) *string {
	return &s
}
