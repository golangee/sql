package ddl

import (
	"fmt"
	"github.com/golangee/reflectplus/src"
	"io"
)

type Table struct {
	doc     string
	name    string
	columns []*Column
	pk      []string
}

func NewTable(name string, cols ...*Column) *Table {
	return &Table{name: name, columns: cols}
}

func (s *Table) Validate() error {
	if !isSafeComment(s.doc) {
		return fmt.Errorf("table '%s' comment is not valid", s.doc)
	}

	if !isSafeName(s.name) {
		return fmt.Errorf("table name '%s' is not a valid identifier", s.name)
	}

	for _, key := range s.pk {
		if !s.hasColumn(key) {
			return fmt.Errorf("primary key '%s' is not defined", key)
		}
	}

	for _, column := range s.columns {
		if err := column.Validate(); err != nil {
			return fmt.Errorf("table '%s': %w", s.name, err)
		}
	}

	return nil
}

func (s *Table) AsGoStruct() *src.TypeBuilder {
	t := src.NewStruct(ddlNameToGoName(s.name)).SetDoc("... is a typed based on the DDL specification of " + s.name + ".\n" + s.doc +
		"\n@ee.sql.Table(\"" + s.name + "\")")
	for _, column := range s.columns {
		t.AddFields(src.NewField(ddlNameToGoName(column.name), typeDeclFromDDLKind(column.kind)).
			AddTag("db", column.name).
			SetDoc("...represents the " + column.name + "(" + string(column.kind) + ") from the DDL specification.\n" + column.doc),
		)
	}
	return t
}

func (s *Table) AsGoMySQLCRUDRepository() *src.TypeBuilder {
	tName := ddlNameToGoName(s.name)
	t := src.NewStruct(tName + "MySQLRepository").SetDoc("...is a CRUD repository based on the DDL specification of " + s.name + "\n" + s.doc)
	if len(s.pk) > 0 {
		t.AddMethods(
			src.NewFunc("FindById").
				SetDoc("...returns the entry identified by its unique primary key.").
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))),
			src.NewFunc("DeleteById").
				SetDoc("...removes the entry identified by its unique primary key.").
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))),
			src.NewFunc("UpdateById").
				SetDoc("...updates an existing entry identified by its unique primary key.").
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))),
			src.NewFunc("Insert").
				SetDoc("...saves a new entry, identified by its unique primary key.").
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))),
		)
	}

	return t
}

func (s *Table) AsMySQL(w io.Writer) (err error) {
	if err = s.Validate(); err != nil {
		return err
	}

	CheckPrintf(w, &err, "CREATE TABLE IF NOT EXISTS `%s` (\n", s.name)

	for i, column := range s.columns {
		CheckPrint(w, &err, " ")
		if err := column.AsMySQL(w); err != nil {
			return fmt.Errorf("invalid column '%s': %w", column.name, err)
		}

		if i < len(s.columns)-1 {
			CheckPrint(w, &err, ",\n")
		}
	}

	if len(s.pk) > 0 {
		CheckPrint(w, &err, ",\n  PRIMARY KEY(")
		for i, key := range s.pk {
			CheckPrintf(w, &err, "`%s`", key)
			if i < len(s.pk)-1 {
				CheckPrint(w, &err, ",")
			}
		}
		CheckPrint(w, &err, ")\n")
	}
	CheckPrint(w, &err, ")")

	if s.doc != "" {
		CheckPrintf(w, &err, " COMMENT '%s'", s.doc)
	}

	return
}

func (s *Table) hasColumn(name string) bool {
	for _, column := range s.columns {
		if column.name == name {
			return true
		}
	}

	return false
}

func (s *Table) AsGraphViz(w io.Writer) (err error) {
	CheckPrintf(w, &err, "graph ER {\n")
	CheckPrintf(w, &err, "node [shape=box];%s;\n", s.name)

	for _, column := range s.columns {
		CheckPrintf(w, &err, "%s -- %s;\n", s.name, column.name)
	}
	CheckPrintf(w, &err, "}\n")

	return
}

func (s *Table) PrimaryKey(names ...string) *Table {
	s.pk = names
	return s
}

func (s *Table) Doc(doc string) *Table {
	s.doc = doc
	return s
}
