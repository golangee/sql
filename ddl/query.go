package ddl

import (
	"github.com/golangee/reflectplus/src"
	"strings"
)

type Dialect int

const (
	SQL Dialect = iota
	MySQL
	PostgreSQL
	Oracle
)

type TupleType int

const (
	AsMultiple TupleType = iota
	AsStruct
	AsArray
)

type BoundedContext struct {
	name    string
	doc     string
	queries []*Query
}

func NewBoundedContext(name string, domain *Domain) *BoundedContext {
	return &BoundedContext{name: name}
}

func (b *BoundedContext) SetDoc(doc string) *BoundedContext {
	b.doc = doc
	return b
}

func (b *BoundedContext) AddQueries(queries ...*Query) *BoundedContext {
	b.queries = append(b.queries, queries...)
	return b
}

type Repository struct {
}

func NewRepository(name string) *Repository {
	return &Repository{}
}

func (r *Repository) Comment(doc string)*Repository{
	return r
}

func (r *Repository) AddQueries(queries...*Query)*Repository{
return nil
}

type Query struct {
	dialect Dialect
	query   string
	name    string
	params  []*src.TypeDecl
	results []*src.TypeDecl
}

func NewQuery(name string, queries ...*RawQuery) *Query {
	return nil
}

func ReadOne(tableName string)*Query{
	return nil
}

func ReadAll(tableName string)*Query{
	return nil
}

func InsertOne(tableName string)*Query{
return nil
}

func DeleteOne(tableName string)*Query{
	return nil
}

func DeleteAll(tableName string)*Query{
	return nil
}

func UpdateOne(tableName string)*Query{
	return nil
}

func CountAll(tableName string)*Query{
	return nil
}

func (q *Query) Input(name string, kind DataType) *Query {
	return q
}

func (q *Query) OutputOne(types ...DataType) *Query {
	return nil
}

func (q *Query) OutputMany(types ...DataType) *Query {
	return nil
}

func (q *Query) Comment(doc string) *Query {
	return nil
}

type RawQuery struct {
	dialect Dialect
	query   string
	name    string
	params  []*src.TypeDecl
	results []*src.TypeDecl
}

// NewRawQuery creates a query which is specialized to the given dialect.
// If a query is cross-sql (e.g. ANSI 99) compatible, one can use the generic
// sql
func NewRawQuery(dialect Dialect, query string) *RawQuery {
	return &RawQuery{dialect: dialect, query: query, name: "name"}
}

func (q *Query) toMySQLExecute() *src.FuncBuilder {
	dbtxIface := src.NewTypeDecl(src.Qualifier(dbCTXInterface().Name())) //TODO full qualified path?

	f := src.NewFunc("...executes the query '"+q.query+"'.").AddParams(
		src.NewParameter("ctx", src.NewTypeDecl("context.Context")),
		src.NewParameter("db", dbtxIface),
	)

	//	f.AddParams(q.params...)
	preparedStmtVarNames := &strings.Builder{}
	for i, param := range q.params {
		_ = param
		preparedStmtVarNames.WriteString("param.Name()")
		if i < len(q.params)-1 {
			preparedStmtVarNames.WriteString(", ")
		}
	}

	f.AddBody(src.NewBlock().
		AddLine(`_, e := db.ExecContext(ctx, "`, q.query, `", `, preparedStmtVarNames.String(), ")").
		Check("e", "failed to execute").
		AddLine("return nil"),
	)

	return f
}

func (q *Query) toMySQLQuery(availableTypes ...*src.TypeBuilder) *src.FuncBuilder {
	dbtxIface := src.NewTypeDecl(src.Qualifier(dbCTXInterface().Name())) //TODO full qualified path?

	f := src.NewFunc("...returns the entries from query '"+q.query+"'.").AddParams(
		src.NewParameter("ctx", src.NewTypeDecl("context.Context")),
		src.NewParameter("db", dbtxIface),
	)

	//f.AddParams(q.params...)
	preparedStmtVarNames := &strings.Builder{}
	for i, param := range q.params {
		_ = param
		preparedStmtVarNames.WriteString("")
		if i < len(q.params)-1 {
			preparedStmtVarNames.WriteString(", ")
		}
	}
	/*
		f.AddBody(src.NewBlock().
			Var("_res", src.NewSliceDecl(structType)).
			AddLine(`_rows, _err := db.Query(ctx, "`, q.query, `", `, preparedStmtVarNames, ")").
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
		)*/

	return f
}
