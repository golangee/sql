package ddl

import (
	"fmt"
	"github.com/golangee/plantuml"
	"github.com/golangee/reflectplus/src"
	"io"
	"strings"
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

func (s *Table) Column(name string) *Column {
	for _, column := range s.columns {
		if column.name == name {
			return column
		}
	}

	return nil
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



func dbCTXInterface() *src.TypeBuilder {
	return src.NewInterface("DBTX").SetDoc("...is the minimal required database access contract.").
		AddMethods(
			src.NewFunc("ExecContext").SetDoc("... executes with context. This is compatible with database/sql.DB or database/sql.Tx.").
				AddParams(
					src.NewParameter("ctx", src.NewTypeDecl("context.Context")),
					src.NewParameter("query", src.NewTypeDecl("string")),
					src.NewParameter("args", src.NewTypeDecl("interface{}")),
				).SetVariadic(true).
				AddResults(
					src.NewParameter("", src.NewTypeDecl("database/sql.Result")),
					src.NewParameter("", src.NewTypeDecl("error")),
				),
			src.NewFunc("QueryContext").SetDoc("... queries with context. This is compatible with database/sql.DB or database/sql.Tx.").
				AddParams(
					src.NewParameter("ctx", src.NewTypeDecl("context.Context")),
					src.NewParameter("query", src.NewTypeDecl("string")),
					src.NewParameter("args", src.NewTypeDecl("interface{}")),
				).SetVariadic(true).
				AddResults(
					src.NewParameter("", src.NewPointerDecl(src.NewTypeDecl("database/sql.Rows"))),
					src.NewParameter("", src.NewTypeDecl("error")),
				),
		)
}
/*
func (s *Table) AsGoMySQLCRUDRepository() *src.TypeBuilder {
	tName := ddlNameToGoName(s.name)
	structType := src.NewTypeDecl(src.Qualifier(ddlNameToGoName(s.name))) //TODO full qualified path?
	dbtxIface := src.NewTypeDecl(src.Qualifier(dbCTXInterface().Name()))  //TODO full qualified path?
	t := src.NewStruct(tName + "MySQLRepository").SetDoc("...is a CRUD repository based on the DDL specification of " + s.name + "\n" + s.doc)
	if len(s.pk) > 0 {
		var ids []*src.Parameter

		varPkParameterNames := ""
		for i, pk := range s.pk {
			column := s.Column(pk)
			varName := pk
			ids = append(ids, src.NewParameter(varName, typeDeclFromDDLKind(column.kind)))
			varPkParameterNames += varName
			if i < len(s.pk)-1 {
				varPkParameterNames += ", "
			}
		}

		t.AddMethods(
			src.NewFunc("FindById").
				SetDoc("...returns the entry identified by its unique primary key.").
				AddParams(
					src.NewParameter("ctx", src.NewTypeDecl("context.Context")),
					src.NewParameter("db", dbtxIface),
				).
				AddParams(ids...).
				AddResults(
					src.NewParameter("", structType),
					src.NewParameter("", src.NewTypeDecl("error"))).
				AddBody(src.NewBlock().
					Var("_res", src.NewSliceDecl(structType)).
					AddLine(`_rows, _err := db.Query("`, s.mysqlFindById(), `", `, varPkParameterNames, ")").
					Check("_err", "failed to query", "_res").
					AddLine("defer _rows.Close()").
					AddLine("for _rows.Next() {").
					AddLine("_entry := ", structType, "{}").
					AddLine("_err := _rows.Scan(", s.goPointer2FieldList("_entry"), ")").
					Check("_err", "failed to scan", "_res").
					AddLine("}").
					AddLine("_err = _rows.Err()").
					Check("_err", "failed to loop rows", "_res").
					AddLine("return _res, nil"),
				),
			src.NewFunc("DeleteById").
				SetDoc("...removes the entry identified by its unique primary key.").
				AddParams(
					src.NewParameter("ctx", src.NewTypeDecl("context.Context")),
					src.NewParameter("db", dbtxIface),
				).
				AddParams(ids...).
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))).
				AddBody(src.NewBlock().
					AddLine(`_, e := db.ExecContext("`, s.mysqlDeleteById(), `", `, varPkParameterNames, ")").
					Check("e", "failed to execute").
					AddLine("return nil"),
				),
			src.NewFunc("UpdateById").
				SetDoc("...updates an existing entry identified by its unique primary key.").
				AddParams(src.NewParameter("v", structType)).
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))),
			src.NewFunc("Insert").
				SetDoc("...saves a new entry, identified by its unique primary key.").
				AddParams(src.NewParameter("v", structType)).
				AddResults(src.NewParameter("", src.NewTypeDecl("error"))),
		)
	}

	return t
}*/

func (s *Table) goPointer2FieldList(varName string) string {
	var names []string
	for _, column := range s.columns {
		names = append(names, "&"+varName+"."+column.name)
	}
	return strings.Join(names, ", ")
}

func (s *Table) mysqlDeleteById() string {
	sb := &strings.Builder{}
	sb.WriteString("DELETE FROM `")
	sb.WriteString(s.name)
	sb.WriteString("` WHERE ")

	for i, pk := range s.pk {
		sb.WriteString("`")
		sb.WriteString(pk)
		sb.WriteString("`")
		sb.WriteString(" = ?")
		if i < len(s.pk)-1 {
			sb.WriteString(" AND ")
		}
	}

	return sb.String()
}

func (s *Table) mysqlFindById() string {
	sb := &strings.Builder{}
	sb.WriteString("SELECT ")
	for i, column := range s.columns {
		sb.WriteString("`")
		sb.WriteString(column.name)
		sb.WriteString("`")
		if i < len(s.columns)-1 {
			sb.WriteString(",")
		}

		sb.WriteString(" ")
	}
	sb.WriteString("FROM `")
	sb.WriteString(s.name)
	sb.WriteString("` WHERE ")

	for i, pk := range s.pk {
		sb.WriteString("`")
		sb.WriteString(pk)
		sb.WriteString("`")
		sb.WriteString(" = ?")
		if i < len(s.pk)-1 {
			sb.WriteString(" AND ")
		}
	}

	return sb.String()
}

func (s *Table) AsMySQL(w io.Writer) (err error) {
	if err = s.Validate(); err != nil {
		return err
	}

	CheckPrintf(w, &err, "CREATE TABLE IF NOT EXISTS `%s` (\n", s.name)

	for i, column := range s.columns {
		CheckPrint(w, &err, " ")
		if err := column.mysql(w); err != nil {
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
	CheckPrint(w, &err, ") ")

	// see https://stackoverflow.com/questions/766809/whats-the-difference-between-utf8-general-ci-and-utf8-unicode-ci/766996#766996
	// https://www.percona.com/live/e17/sites/default/files/slides/Collations%20in%20MySQL%208.0.pdf
	//
	// we enforce correct unicode support for mysql and index/sorting collations. For mysql 8.0 using
	// accent insensitive/case insensitive Unicode 9 support utf8mb4_0900_ai_ci would be better but not compatible
	// with mariadb, so we use a fixed older version for reproducibility across different database servers.
	CheckPrint(w, &err, "CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci")

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

/*
func (s *Table) AsPlantUML(w io.Writer) (err error) {

	myStruct := type2Uml(s.AsGoStruct())

	diagram := plantuml.NewDiagram().Add(myStruct, type2Uml(s.AsGoMySQLCRUDRepository()).Uses(myStruct.Name()))
	CheckPrint(w, &err, plantuml.String(diagram))
	return
}
 */

func type2Uml(myType *src.TypeBuilder) *plantuml.Class {
	res := plantuml.NewClass(myType.Name())
	if myType.Doc() != "" {
		res.NoteTop(plantuml.NewNote(myType.Doc()))
	}

	for _, field := range myType.Fields() {
		res.AddAttrs(plantuml.Attr{
			Visibility: plantuml.Public,
			Abstract:   false,
			Static:     false,
			Name:       field.Name(),
			Type:       decl2str(field.Type()),
		})
	}

	for _, method := range myType.Methods() {
		res.AddAttrs(plantuml.Attr{
			Visibility: plantuml.Public,
			Abstract:   false,
			Static:     false,
			Name:       method.Name() + "(" + params2str(method.Params()) + ")",
			Type:       params2str(method.Results()),
		})
	}

	return res
}

func decl2str(d *src.TypeDecl) string {
	b := &src.BufferedWriter{}
	d.Emit(b)
	return b.String()
}

func params2str(params []*src.Parameter) string {
	sb := &strings.Builder{}
	if len(params) > 1 {
		sb.WriteString("(")
	}
	for i, param := range params {
		b := &src.BufferedWriter{}
		param.Decl().Emit(b)
		sb.WriteString(strings.TrimSpace(b.String()))
		if i < len(params)-1 {
			sb.WriteString(", ")
		}
	}
	if len(params) > 1 {
		sb.WriteString(")")
	}
	return sb.String()
}

func (s *Table) PrimaryKey(names ...string) *Table {
	s.pk = names
	return s
}

func (s *Table) Comment(doc string) *Table {
	s.doc = doc
	return s
}
