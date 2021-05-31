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

package ddl

import (
	"fmt"
)

// ParseResult contains all information that could be parsed from the SQL.
type ParseResult struct {
	// Tables are all parsed CREATE TABLE statements.
	Tables []Table
	// AlterStatements are all parsed ALTER TABLE statements.
	AlterStatements []AlterStatement
}

// Table represents a CREATE definition for a single SQL table.
type Table struct {
	Name        string
	IfNotExists bool
	Columns     []Column
	ForeignKeys []ForeignKeyConstraint
	Keys        []Key
}

// A Column defined in a Table.
type Column struct {
	Name       string
	Type       string
	NotNull    bool
	PrimaryKey bool
	Unique     bool
	Default    *string
}

// ForeignKeyConstraint is a FOREIGN KEY constraint in SQL.
type ForeignKeyConstraint struct {
	Name            *string
	Column          string
	ReferenceTable  string
	ReferenceColumn string
}

// Key is an SQL INDEX.
type Key struct {
	// Name is the name of this index. Might be nil if it has no name.
	Name *string
	// OnColumn is the name of the column, this index applies to.
	OnColumn string
}

// AlterAddColumn represents an ALTER TABLE 'Table' ADD COLUMN statement.
type AlterAddColumn struct {
	// Table is the name of the table to which the column is added.
	Table string
	// Column is the column to add.
	Column Column
	// First is true when the column should be added to the front of the table.
	First bool
	// After is set when the column should be added after the given column name in the table.
	After *string
}

// AlterDropColumn describes an ALTER TABLE 'table' DROP COLUMN 'column'.
type AlterDropColumn struct {
	// Table is the name of the table from which the column is removed.
	Table string
	// Column is the name of the column that will be removed.
	Column string
}

// AlterAddIndex describes a CREATE INDEX 'name' ON 'table' ('column') statement.
type AlterAddIndex struct {
	// Table is the name of the table to which the statement is added.
	Table string
	// Name is the name of the new index.
	Name string
	// Column is the name of the column the index will be applied to.
	Column string
	// Unique is set, if the index is UNIQUE.
	Unique bool
}

// AlterDropIndex describes a DROP INDEX 'name' ON 'table' or a ALTER TABLE 'table' DROP INDEX 'index' statement.
type AlterDropIndex struct {
	// Table is the name of the table from which the index will be removed.
	Table string
	// Index is the name of the index that should be removed.
	Index string
}

// AlterStatement might be ADD COLUMN, DROP COLUMN, ADD INDEX, DROP INDEX.
// It can be applied to a table to perform the corresponding operation.
type AlterStatement interface {
	// TableName returns the name of the Table that this statement wants to modify.
	TableName() string
	// ApplyTo applies the alteration to the given table.
	// Returns an error if applying failed.
	ApplyTo(table *Table) error
}

// DropErrorNotFound signifies that a certain thing could not be found.
type DropErrorNotFound struct {
	property string
}

func (e DropErrorNotFound) Error() string {
	return fmt.Sprintf("%s could not be dropped because it was not present", e.property)
}

func (a AlterAddColumn) TableName() string {
	return a.Table
}

func (a AlterAddColumn) ApplyTo(table *Table) error {
	insertAt := len(table.Columns)

	if a.First {
		insertAt = 0
	}

	if a.After != nil {
		for i, col := range table.Columns {
			if col.Name == *a.After {
				insertAt = i + 1
			}
		}
	}

	if len(table.Columns) == insertAt {
		// Append
		table.Columns = append(table.Columns, a.Column)
	} else {
		// Index points somewhere in the middle
		table.Columns = append(table.Columns[:insertAt+1], table.Columns[insertAt:]...)
		table.Columns[insertAt] = a.Column
	}

	return nil
}

func (a AlterDropColumn) TableName() string {
	return a.Table
}

func (a AlterDropColumn) ApplyTo(table *Table) error {
	index := -1

	for i, col := range table.Columns {
		if col.Name == a.Column {
			index = i

			break
		}
	}

	if index < 0 {
		return DropErrorNotFound{a.Column}
	}

	table.Columns = append(table.Columns[:index], table.Columns[index+1:]...)

	return nil
}

func (a AlterAddIndex) TableName() string {
	return a.Table
}

func (a AlterAddIndex) ApplyTo(table *Table) error {
	table.Keys = append(table.Keys, Key{
		Name:     &a.Name,
		OnColumn: a.Column,
	})

	return nil
}

func (a AlterDropIndex) TableName() string {
	return a.Table
}

func (a AlterDropIndex) ApplyTo(table *Table) error {
	index := -1

	for i, key := range table.Keys {
		if key.Name != nil && *key.Name == a.Index {
			index = i

			break
		}
	}

	if index < 0 {
		return DropErrorNotFound{a.Index}
	}

	table.Keys = append(table.Keys[:index], table.Keys[index+1:]...)

	return nil
}
