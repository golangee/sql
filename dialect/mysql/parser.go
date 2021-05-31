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

package mysql

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/golangee/sql/ddl"
	"github.com/golangee/sql/dialect/mysql/parser"
	"strings"
)

// Parse extracts all tables from CREATE TABLE statements from a given set of SQL statements.
func Parse(sql string) (*ddl.ParseResult, error) {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewMySqlLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	parser := parser.NewMySqlParser(stream)

	parser.RemoveErrorListeners()
	lexer.RemoveErrorListeners()

	errorCollector := &errorCollector{}
	parser.AddErrorListener(errorCollector)
	lexer.AddErrorListener(errorCollector)

	listener := newListener()
	antlr.ParseTreeWalkerDefault.Walk(listener, parser.Root())

	if len(errorCollector.errors.messages) > 0 {
		return nil, errorCollector.errors
	}

	return &ddl.ParseResult{
		Tables:          listener.Tables,
		AlterStatements: listener.AlterStatements,
	}, nil
}

// errorCollector collects all errors that occur during parsing.
type errorCollector struct {
	*antlr.DefaultErrorListener
	errors SyntaxErrors
}

type SyntaxErrors struct {
	messages []string
}

func (s SyntaxErrors) Error() string {
	return "Errors: " + strings.Join(s.messages, "; ")
}

func (c *errorCollector) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{},
	line, column int, msg string, e antlr.RecognitionException) {
	c.errors.messages = append(c.errors.messages, msg) // More information may be useful
}

type listener struct {
	*parser.BaseMySqlParserListener
	// The table that is currently being parsed
	BuildingTable *ddl.Table
	// The column that is currently being parsed
	BuildingColumn *ddl.Column
	// The constraint that is currently parsed
	BuildingForeignKeyConstraint *ddl.ForeignKeyConstraint
	// A list of all parsed CREATE TABLE statements
	Tables []ddl.Table
	// A list of parsed ALTER TABLE statements
	AlterStatements []ddl.AlterStatement
}

func newListener() *listener {
	return &listener{}
}

// Trim SQL names for columns and tables by stripping quotes ("') and backticks (`).
func trimName(name string) string {
	return strings.Trim(name, "`'\"")
}

// A new CREATE TABLE statement was detected.
func (l *listener) EnterColumnCreateTable(ctx *parser.ColumnCreateTableContext) {
	name := ctx.TableName().GetText()
	name = trimName(name)
	l.BuildingTable = &ddl.Table{
		Name:        name,
		IfNotExists: ctx.IfNotExists() != nil,
	}
}

// A CREATE TABLE statement is done processing.
// Append the table to the list of parsed ones.
func (l *listener) ExitColumnCreateTable(ctx *parser.ColumnCreateTableContext) {
	l.Tables = append(l.Tables, *l.BuildingTable)
	l.BuildingTable = nil
}

// --- Column specific callbacks

// A new column declaration is visited.
func (l *listener) EnterColumnDeclaration(ctx *parser.ColumnDeclarationContext) {
	l.BuildingColumn = &ddl.Column{}
}

// The column declaration is finished, save it.
func (l *listener) ExitColumnDeclaration(ctx *parser.ColumnDeclarationContext) {
	l.BuildingTable.Columns = append(l.BuildingTable.Columns, *l.BuildingColumn)
	l.BuildingColumn = nil
}

func (l *listener) EnterUid(ctx *parser.UidContext) { //nolint
	if l.BuildingColumn != nil {
		if len(l.BuildingColumn.Name) == 0 {
			l.BuildingColumn.Name = trimName(ctx.GetText())
		}
	}
}

func (l *listener) EnterDataType(ctx *parser.DataTypeContext) {
	if l.BuildingColumn != nil {
		l.BuildingColumn.Type = ctx.GetText()
	}
}

// --- Callbacks for ALTER TABLE statements

// An ALTER TABLE statement was detected. Prepare the table name, so that it is available
// for saving the smaller statements.
func (l *listener) EnterAlterTable(ctx *parser.AlterTableContext) {
	tableName := ctx.TableName().GetText()
	tableName = trimName(tableName)
	l.BuildingTable = &ddl.Table{Name: tableName}
}

// An ALTER TABLE statement was parsed, reset the table.
func (l *listener) ExitAlterTable(ctx *parser.AlterTableContext) {
	l.BuildingTable = nil
}

// Prepare a new ADD COLUMN statement.
func (l *listener) EnterAlterByAddColumn(ctx *parser.AlterByAddColumnContext) {
	l.BuildingColumn = &ddl.Column{}
}

// We parsed an ADD COLUMN statement. Save it.
func (l *listener) ExitAlterByAddColumn(ctx *parser.AlterByAddColumnContext) {
	addStatement := ddl.AlterAddColumn{
		Table:  l.BuildingTable.Name,
		Column: *l.BuildingColumn,
	}

	if ctx.AFTER() != nil {
		afterColumn := ctx.Uid(1).GetText()
		afterColumn = trimName(afterColumn)
		addStatement.After = &afterColumn
	}

	if ctx.FIRST() != nil {
		addStatement.First = true
	}

	l.AlterStatements = append(l.AlterStatements, addStatement)
	l.BuildingColumn = nil
}

// We parsed a DROP COLUMN statement. Save it.
func (l *listener) ExitAlterByDropColumn(ctx *parser.AlterByDropColumnContext) {
	l.AlterStatements = append(l.AlterStatements, ddl.AlterDropColumn{
		Table:  l.BuildingTable.Name,
		Column: trimName(ctx.Uid().GetText()),
	})
}

func (l *listener) EnterCreateIndex(ctx *parser.CreateIndexContext) {
	indexName := ctx.Uid().GetText()
	indexName = trimName(indexName)
	onTableName := ctx.TableName().GetText()
	onTableName = trimName(onTableName)

	// An index can be created on many columns. To keep things simple, we will only support one column.
	// This is analogous to the limitations for FOREIGN KEY constraints. See EnterForeignKeyTableConstraint.
	columnName := ctx.IndexColumnNames().GetText()
	columnName = trimName(strings.Trim(columnName, "()"))

	l.AlterStatements = append(l.AlterStatements, ddl.AlterAddIndex{
		Table:  onTableName,
		Name:   indexName,
		Column: columnName,
		Unique: ctx.UNIQUE() != nil,
	})
}

// A DROP INDEX 'index' ON 'table' statement was parsed.
func (l *listener) EnterDropIndex(ctx *parser.DropIndexContext) {
	indexName := ctx.Uid().GetText()
	indexName = trimName(indexName)
	onTableName := ctx.TableName().GetText()
	onTableName = trimName(onTableName)

	l.AlterStatements = append(l.AlterStatements, ddl.AlterDropIndex{
		Table: onTableName,
		Index: indexName,
	})
}

// A ALTER TABLE 'table' DROP INDEX 'index' statement was parsed.
func (l *listener) EnterAlterByDropIndex(ctx *parser.AlterByDropIndexContext) {
	indexName := ctx.Uid().GetText()
	indexName = trimName(indexName)
	onTableName := l.BuildingTable.Name
	l.AlterStatements = append(l.AlterStatements, ddl.AlterDropIndex{
		Table: onTableName,
		Index: indexName,
	})
}

// --- Callbacks for building constraints

// A FOREIGN KEY is visited.
func (l *listener) EnterForeignKeyTableConstraint(ctx *parser.ForeignKeyTableConstraintContext) {
	l.BuildingForeignKeyConstraint = &ddl.ForeignKeyConstraint{}

	if ctx.GetName() != nil {
		constraintName := trimName(ctx.GetName().GetText())
		l.BuildingForeignKeyConstraint.Name = &constraintName
	}

	// A FOREIGN KEY constraint can reference multiple columns ("composite key").
	// For the sake of keeping it simple, we will assume that only a single column is specified.
	column := ctx.IndexColumnNames().GetText()
	column = trimName(strings.Trim(column, "()"))
	l.BuildingForeignKeyConstraint.Column = column
}

// We can get the names of what a FOREIGN KEY is referencing here.
func (l *listener) EnterReferenceDefinition(ctx *parser.ReferenceDefinitionContext) {
	if l.BuildingForeignKeyConstraint != nil {
		l.BuildingForeignKeyConstraint.ReferenceTable = trimName(ctx.TableName().GetText())
		// FOREIGN KEYs can be composite. See above on why we ignore that.
		column := ctx.IndexColumnNames().GetText()
		column = trimName(strings.Trim(column, "()"))
		l.BuildingForeignKeyConstraint.ReferenceColumn = column
		l.BuildingTable.ForeignKeys = append(l.BuildingTable.ForeignKeys, *l.BuildingForeignKeyConstraint)
		l.BuildingForeignKeyConstraint = nil
	}
}

// NOT NULL constraint.
func (l *listener) EnterNullColumnConstraint(ctx *parser.NullColumnConstraintContext) {
	if l.BuildingColumn != nil {
		l.BuildingColumn.NotNull = true
	}
}

// PRIMARY KEY constraint.
func (l *listener) EnterPrimaryKeyColumnConstraint(ctx *parser.PrimaryKeyColumnConstraintContext) {
	if l.BuildingColumn != nil {
		l.BuildingColumn.PrimaryKey = true
	}
}

// UNIQUE constraint.
func (l *listener) EnterUniqueKeyColumnConstraint(ctx *parser.UniqueKeyColumnConstraintContext) {
	if l.BuildingColumn != nil {
		l.BuildingColumn.Unique = true
	}
}

// DEFAULT constraint.
func (l *listener) EnterDefaultColumnConstraint(ctx *parser.DefaultColumnConstraintContext) {
	if l.BuildingColumn != nil {
		defaultValue := ctx.DefaultValue().GetText()
		l.BuildingColumn.Default = &defaultValue
	}
}

// A KEY constraint, which represents an index an a column.
func (l *listener) EnterSimpleIndexDeclaration(ctx *parser.SimpleIndexDeclarationContext) {
	key := ddl.Key{}

	if ctx.Uid() != nil {
		keyName := ctx.Uid().GetText()
		keyName = trimName(keyName)
		key.Name = &keyName
	}

	// An index can be created on many columns. To keep things simple, we will only support one column.
	// This is analogous to the limitations for FOREIGN KEY constraints. See EnterForeignKeyTableConstraint.
	columnName := ctx.IndexColumnNames().GetText()
	columnName = trimName(strings.Trim(columnName, "()"))
	key.OnColumn = columnName

	l.BuildingTable.Keys = append(l.BuildingTable.Keys, key)
}
