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

package ddl_test

import (
	"github.com/golangee/sql/ddl"
	"testing"
)

func TestAlterAddColumn_Apply(t *testing.T) {
	table := ddl.Table{
		Name:    "TestTable",
		Columns: []ddl.Column{},
	}

	alter := ddl.AlterAddColumn{
		Column: ddl.Column{Name: "A"},
	}
	if err := alter.ApplyTo(&table); err != nil {
		t.Error(err)
	}

	if len(table.Columns) < 1 || table.Columns[0].Name != "A" {
		t.Fatalf("Failed to add column to empty table")
	}

	alter = ddl.AlterAddColumn{
		Column: ddl.Column{Name: "B"},
		First:  true,
	}
	if err := alter.ApplyTo(&table); err != nil {
		t.Error(err)
	}

	if len(table.Columns) < 1 || table.Columns[0].Name != "B" {
		t.Fatalf("Failed to add column first")
	}

	b := "B"
	alter = ddl.AlterAddColumn{
		Column: ddl.Column{Name: "C"},
		After:  &b,
	}

	if err := alter.ApplyTo(&table); err != nil {
		t.Error(err)
	}

	if len(table.Columns) < 2 || table.Columns[1].Name != "C" {
		t.Fatalf("Failed to add column to empty table")
	}
}

func TestAlterAddIndex_Apply(t *testing.T) {
	table := ddl.Table{}
	if err := (ddl.AlterAddIndex{Column: "A"}.ApplyTo(&table)); err != nil {
		t.Fatal(err)
	}

	if len(table.Keys) < 1 || table.Keys[0].OnColumn != "A" {
		t.Fatalf("Failed to insert key")
	}
}

func TestAlterDropColumn_Apply(t *testing.T) {
	table := ddl.Table{
		Columns: []ddl.Column{
			{Name: "A"},
			{Name: "B"},
			{Name: "C"},
			{Name: "D"},
			{Name: "E"},
		},
	}
	if err := (ddl.AlterDropColumn{Column: "A"}.ApplyTo(&table)); err != nil {
		t.Fatal(err)
	}

	if err := (ddl.AlterDropColumn{Column: "C"}.ApplyTo(&table)); err != nil {
		t.Fatal(err)
	}

	if err := (ddl.AlterDropColumn{Column: "E"}.ApplyTo(&table)); err != nil {
		t.Fatal(err)
	}

	if len(table.Columns) != 2 {
		t.Fatalf("Failed to drop a column")
	}
}

func TestAlterDropIndex_Apply(t *testing.T) {
	indexName := "idx"
	table := ddl.Table{
		Keys: []ddl.Key{
			{OnColumn: "A", Name: &indexName},
		},
	}

	if err := (ddl.AlterDropIndex{Index: indexName}.ApplyTo(&table)); err != nil {
		t.Fatal(err)
	}

	if len(table.Keys) > 0 {
		t.Fatalf("Failed to drop key")
	}
}
