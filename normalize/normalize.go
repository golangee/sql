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

package normalize

import (
	"fmt"
	"sort"
	"sql/model"
)

func Tables(tables []model.Table) string {
	// Sort tables by name
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	result := ""

	for _, table := range tables {
		result += Table(table)
	}

	return result
}

func Table(table model.Table) string {
	result := "CREATE TABLE"
	if table.IfNotExists {
		result += " IF NOT EXISTS"
	}
	// Assemble column declarations and constraints as the statements body.
	body := Columns(table.Columns)
	if len(table.ForeignKeys) > 0 {
		body += "," + ForeignKeys(table.ForeignKeys)
	}

	if len(table.Keys) > 0 {
		body += "," + Keys(table.Keys)
	}

	result += fmt.Sprintf(" `%s` (%s);", table.Name, body)

	return result
}

func Columns(columns []model.Column) string {
	// Sort columns by name
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Name < columns[j].Name
	})

	result := ""

	for i, column := range columns {
		if i > 0 {
			result += ","
		}

		result += Column(column)
	}

	return result
}

func Column(column model.Column) string {
	result := fmt.Sprintf("`%s` %s", column.Name, column.Type)

	// Append constraints alphabetically

	if column.Default != nil {
		result += " DEFAULT " + *column.Default
	}

	if column.NotNull {
		result += " NOT NULL"
	}

	if column.PrimaryKey {
		result += " PRIMARY KEY"
	}

	if column.Unique {
		result += " UNIQUE"
	}

	return result
}

func ForeignKeys(keys []model.ForeignKeyConstraint) string {
	// Sort keys by constraint name then by the column they apply to.
	// This is achieved by building a string for comparison that has the format 'constraint.column'
	sort.Slice(keys, func(i, j int) bool {
		keyI := fmt.Sprintf("%s.%s", nilString(keys[i].Name), keys[i].Column)
		keyJ := fmt.Sprintf("%s.%s", nilString(keys[j].Name), keys[j].Column)

		return keyI < keyJ
	})

	result := ""

	for i, key := range keys {
		if i > 0 {
			result += ","
		}

		result += ForeignKey(key)
	}

	return result
}

func ForeignKey(key model.ForeignKeyConstraint) string {
	result := ""
	if key.Name != nil {
		result += fmt.Sprintf("CONSTRAINT %s ", *key.Name)
	}

	result += fmt.Sprintf("FOREIGN KEY (`%s`) REFERENCES `%s`(`%s`)", key.Column, key.ReferenceTable, key.ReferenceColumn)

	return result
}

func Keys(keys []model.Key) string {
	// Sort keys by constraint name then by the column they apply to.
	// This is achieved by building a string for comparison that has the format 'constraint.column'
	sort.Slice(keys, func(i, j int) bool {
		keyI := fmt.Sprintf("%s.%s", nilString(keys[i].Name), keys[i].OnColumn)
		keyJ := fmt.Sprintf("%s.%s", nilString(keys[j].Name), keys[j].OnColumn)

		return keyI < keyJ
	})

	result := ""

	for i, key := range keys {
		if i > 0 {
			result += ","
		}

		result += Key(key)
	}

	return result
}

func Key(key model.Key) string {
	result := "KEY"
	if key.Name != nil {
		result += fmt.Sprintf(" `%s`", *key.Name)
	}

	result += fmt.Sprintf("(`%s`)", key.OnColumn)

	return result
}

func AlterStatements(alterStatements []model.AlterStatement) string {
	// No sorting or anything is allowed here, as that would change the meaning!
	result := ""

	for _, stmt := range alterStatements {
		result += AlterTableStatement(stmt)
	}

	return result
}

func AlterTableStatement(alterStatement model.AlterStatement) string {
	switch stmt := alterStatement.(type) {
	case model.AlterAddColumn:
		return AlterAddColumn(stmt)
	case model.AlterDropColumn:
		return AlterDropColumn(stmt)
	case model.AlterAddIndex:
		return AlterAddIndex(stmt)
	case model.AlterDropIndex:
		return AlterDropIndex(stmt)
	default:
		return "not implemented"
	}
}

func AlterAddColumn(add model.AlterAddColumn) string {
	result := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s", add.Table, Column(add.Column))
	if add.First {
		result += " FIRST"
	} else if add.After != nil {
		result += fmt.Sprintf(" AFTER `%s`", *add.After)
	}

	return result + ";"
}

func AlterDropColumn(drop model.AlterDropColumn) string {
	return fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`;", drop.Table, drop.Column)
}

func AlterAddIndex(index model.AlterAddIndex) string {
	pre := "CREATE INDEX"
	if index.Unique {
		pre = "CREATE UNIQUE INDEX"
	}

	return fmt.Sprintf("%s `%s` ON `%s`(`%s`);", pre, index.Name, index.Table, index.Column)
}

func AlterDropIndex(drop model.AlterDropIndex) string {
	return fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", drop.Table, drop.Index)
}

// Interpret nil as an empty string.
func nilString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
