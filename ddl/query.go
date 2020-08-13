package ddl

import (
	"github.com/golangee/reflectplus/src"
	"strings"
)

type Dialect int

const (
	DialectUnknown Dialect = iota
	DialectMySQL
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

type Query struct {
	dialect Dialect
	query   string
	name    string
	params  []*src.TypeDecl
	results []*src.TypeDecl
}

func NewRawQuery(dialect Dialect, name, query string) *Query {
	return &Query{dialect: dialect, query: query, name: name}
}

func (q *Query) StructParams(types ...*src.TypeDecl) *Query {
	return q
}

func (q *Query) Params(types ...*src.TypeDecl) *Query {
	q.params = append(q.params, types...)
	return q
}

func (q *Query) StructResult(fields ...*src.TypeDecl) *Query {
	q.results = append(q.results, fields...)
	return q
}

func (q *Query) StructResults(fields ...*src.TypeDecl) *Query {
	q.results = append(q.results, fields...)
	return q
}

func (q *Query) Results(types *src.TypeDecl) *Query {
	q.results = append(q.results, types)
	return q
}

func (q *Query) MapResult(types ...*src.TypeDecl) *Query {
	q.results = append(q.results, types...)
	return q
}

func (q *Query) BindParam(name string) *Query {
	return q
}

func (q *Query) toMySQLFunc(availableTypes ...*src.TypeBuilder) *src.FuncBuilder {
	if len(q.results) == 0 {
		return q.toMySQLExecute()
	}

	return q.toMySQLQuery(availableTypes...)
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
