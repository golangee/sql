// Copyright 2020 Torben Schinke
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

package sql

import (
	"context"
	"fmt"
	"github.com/golangee/reflectplus"
	"reflect"
	"strings"
)

var ErrEntityNotFound = fmt.Errorf("entity not found")

type sqlResultType int

const (
	// returns error
	Stmt sqlResultType = 0

	// returns []T,error
	List sqlResultType = 1

	// returns T,error
	Single sqlResultType = 2
)

// MustMakeSQLRepositories tries to autodetect and implement all interfaces marked with @ee.stereotype.Repository.
func MustMakeSQLRepositories(dbtx DBTX) []interface{} {
	dialect, err := DetectDialect(dbtx)
	if err != nil {
		panic(err)
	}
	repos, err := MakeSQLRepositories(dialect)
	if err != nil {
		panic(err)
	}
	return repos
}

// MakeSQLRepositories scans for @ee.stereotype.Repository annotated interfaces and tries to create a proxy instance for each.
func MakeSQLRepositories(dialect Dialect) ([]interface{}, error) {
	var res []interface{}
	for _, iface := range reflectplus.Interfaces() {
		if reflectplus.Annotations(iface.Annotations).Has(AnnotationRepository) {
			repo, err := NewRepository(dialect, iface)
			if err != nil {
				return nil, err
			}
			proxy, err := reflectplus.NewProxy(iface.ImportPath, iface.Name, repo.HandleQuery)
			if err != nil {
				return nil, err
			}
			res = append(res, proxy)
		}
	}
	return res, nil
}

type Repository struct {
	methods map[string]reflectplus.InvocationHandler
}

// NewRepository interprets the given interface as an @ee.stereotype.Repository and tries to create a generic Repository
// with a reflectplus-based handler for it.
func NewRepository(dialect Dialect, iface reflectplus.Interface) (*Repository, error) {
	r := &Repository{
		methods: make(map[string]reflectplus.InvocationHandler),
	}
	for _, method := range iface.Methods {
		annotations := filterByDialect(dialect, method.GetAnnotations().FindAll(AnnotationQuery))
		if len(annotations) != 1 {
			return nil, reflectplus.PositionalError(method, fmt.Errorf("each method must have exactly one '%s' annotation", AnnotationQuery))
		}
		annotation := annotations[0]

		err := validateFirstParamCtx(method)
		if err != nil {
			return nil, reflectplus.PositionalError(method, err)
		}

		resType, err := validateResultType(method)
		if err != nil {
			return nil, reflectplus.PositionalError(method, err)
		}

		query := annotation.Value()
		if len(query) == 0 {
			return nil, reflectplus.PositionalError(method, fmt.Errorf("query must not be empty"))
		}
		stmt, err := NamedParameterStatement(query).Prepare(dialect, paramNames(method)[1:])
		if err != nil {
			return nil, reflectplus.PositionalError(annotation, err)
		}
		switch resType {
		case Stmt:
			r.methods[method.Name] = func(method string, args ...interface{}) []interface{} {
				ctx := args[0].(context.Context)
				args = args[1:]
				err := execStmt(ctx, stmt, args...)
				if err != nil {
					err = fmt.Errorf("statement \"%s\" failed: %w", stmt, err)
				}
				return []interface{}{err}
			}
		case List:
			retSliceType := method.Returns[0].Type.Params[0]
			resType := reflectplus.FindType(retSliceType.ImportPath, retSliceType.Identifier)
			if resType == nil {
				return nil, reflectplus.PositionalError(method, fmt.Errorf("cannot resolve slice type %s#%s", retSliceType.ImportPath, retSliceType.Identifier))
			}
			strct := reflectplus.FindStruct(retSliceType.ImportPath, retSliceType.Identifier)
			if strct == nil {
				return nil, reflectplus.PositionalError(method, fmt.Errorf("cannot resolve slice type as struct %s#%s", retSliceType.ImportPath, retSliceType.Identifier))
			}

			r.methods[method.Name] = func(methodName string, args ...interface{}) []interface{} {
				ctx := args[0].(context.Context)
				args = args[1:]

				list, err := execQueryList(ctx, stmt, *strct, resType, args...)
				if err != nil {
					err = fmt.Errorf("statement \"%s\" failed: %w", stmt, err)
				}
				return []interface{}{list, err}
			}
		case Single:
			retType := method.Returns[0].Type
			resType := reflectplus.FindType(retType.ImportPath, retType.Identifier)
			if resType == nil {
				return nil, reflectplus.PositionalError(method, fmt.Errorf("cannot resolve return type %s#%s", retType.ImportPath, retType.Identifier))
			}
			strct := reflectplus.FindStruct(retType.ImportPath, retType.Identifier)
			if strct == nil {
				return nil, reflectplus.PositionalError(method, fmt.Errorf("cannot resolve slice type as struct %s#%s", retType.ImportPath, retType.Identifier))
			}

			r.methods[method.Name] = func(method string, args ...interface{}) []interface{} {
				ctx := args[0].(context.Context)
				args = args[1:]

				entry, err := execQuerySingle(ctx, stmt, *strct, resType, args...)
				if err != nil {
					err = fmt.Errorf("statement \"%s\" failed: %w", stmt, err)
				}
				return []interface{}{entry, err}
			}
		default:
			panic("not implemented")
		}
	}
	return r, nil
}

func paramNames(method reflectplus.Method) []string {
	var res []string
	for _, p := range method.Params {
		res = append(res, p.Name)
	}
	return res
}

func execStmt(ctx context.Context, stmt DialectStatement, args ...interface{}) error {
	tx, err := FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(tx, ctx, args...)
	if err != nil {
		return err
	}
	return err
}

func execQuerySingle(ctx context.Context, stmt DialectStatement, strct reflectplus.Struct, resType reflect.Type, args ...interface{}) (interface{}, error) {
	tx, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(tx, ctx, args...)
	if err != nil {
		return nil, err
	}
	colNames, _ := rows.Columns()
	tmpPtrInstance := reflect.New(resType)
	cols, err := structFieldPointersToSlice(colNames, tmpPtrInstance.Interface())
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	found := false
	for rows.Next() {
		err := rows.Scan(cols...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		found = true
		break
		//fmt.Printf("%v\n", tmpPtrInstance.Elem())
	}
	err = rows.Err()
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	if !found {
		return tmpPtrInstance.Elem().Interface(), ErrEntityNotFound
	}
	return tmpPtrInstance.Elem().Interface(), nil
}

func execQueryList(ctx context.Context, stmt DialectStatement, strct reflectplus.Struct, resType reflect.Type, args ...interface{}) (interface{}, error) {
	tx, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	slice := reflect.MakeSlice(reflect.SliceOf(resType), 0, 10)
	rows, err := stmt.QueryContext(tx, ctx, args...)
	if err != nil {
		return nil, err
	}
	colNames, _ := rows.Columns()
	tmpPtrInstance := reflect.New(resType)
	cols, err := structFieldPointersToSlice(colNames, tmpPtrInstance.Interface())
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(cols...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		slice = reflect.Append(slice, tmpPtrInstance.Elem())
		//fmt.Printf("%v\n", tmpPtrInstance.Elem())
	}
	err = rows.Err()
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	return slice.Interface(), nil
}

// This is stupid slow, so we may optimize it using pointer offsets of fields and pointers into a slice of values.
// However for large sets probably not so important, because this is not used in the actual hot-path. Also we better
// keep that n^2 loop, which is probably still faster than a map allocation.
//
// Or we throw that away and just generate source code for that, which is probably more idiomatic
func structFieldPointersToSlice(selectedNames []string, u interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(u).Elem()
	typ := reflect.Indirect(val).Type()
	v := make([]interface{}, len(selectedNames))
	for _, colName := range selectedNames {
		found := false
		for i := 0; i < val.NumField(); i++ {
			name := typ.Field(i).Tag.Get(AnnotationName)
			if len(name) == 0 {
				name = typ.Field(0).Name
			}
			if colName == name {
				valueField := val.Field(i)
				v[i] = valueField.Addr().Interface()
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("type '%s' has not a field named '%s'. Available columns are (%s). Either change your query or use a struct tag like `%s:\"%s\"`", typ.Name(), colName, strings.Join(selectedNames, ","), AnnotationName, colName)
		}
	}

	return v, nil
}

func validateFirstParamCtx(method reflectplus.Method) error {
	if len(method.Params) < 1 {
		return fmt.Errorf("a method must have at least a (context.Context) parameter")
	}
	if !(method.Params[0].Type.ImportPath == "context" && method.Params[0].Type.Identifier == "Context") {
		return fmt.Errorf("the first parameter must be (context.Context)")
	}
	return nil
}

func validateResultType(method reflectplus.Method) (sqlResultType, error) {
	if !(len(method.Returns) == 1 || len(method.Returns) == 2) {
		return -1, fmt.Errorf("invalid return parameters, must be either (error) or (<T>,error)")
	}

	if len(method.Returns) == 1 {
		if !(method.Returns[0].Type.Identifier == "error" && method.Returns[0].Type.ImportPath == "") {
			return -1, fmt.Errorf("invalid return parameter, must be (error)")
		}
		return Stmt, nil
	}

	if !(method.Returns[1].Type.Identifier == "error" && method.Returns[1].Type.ImportPath == "") {
		return -1, fmt.Errorf("invalid return parameters, must be (<T>,error)")
	}

	if method.Returns[0].Type.Stars > 0 {
		return -1, fmt.Errorf("invalid return parameters, (<T>,error) but T must not be a pointer or a slice of pointers")
	}

	if method.Returns[0].Type.Identifier != "[]" {
		return Single, nil
	}

	if method.Returns[0].Type.Identifier == "[]" && method.Returns[0].Type.Params[0].Stars > 0 {
		return -1, fmt.Errorf("invalid return parameters, ([]<T>,error) but T must not be a pointer")
	}
	return List, nil
}

func (r *Repository) HandleQuery(method string, args ...interface{}) []interface{} {
	return r.methods[method](method, args...)
}

// filterByDialect removes any annotation which does not match the given dialect.
// It only removes generic ee.sql.Query if another with that dialect was found.
func filterByDialect(dialect Dialect, annotations []reflectplus.Annotation) []reflectplus.Annotation {
	var lastQuery reflectplus.Annotation
	var res []reflectplus.Annotation
	for _, a := range annotations {
		if a.Name == AnnotationQuery {
			if lastQuery.Name == "" && (a.AsString(DialectValue) == "" || dialect.Matches(a.AsString(DialectValue))) {
				lastQuery = a
			}
			if dialect.Matches(a.AsString(DialectValue)) {
				lastQuery = a
			}
		} else {
			if a.AsString(DialectValue) == "" || dialect.Matches(a.AsString(DialectValue)) {
				res = append(res, a)
			}
		}
	}
	if lastQuery.Name != "" {
		res = append(res, lastQuery)
	}
	return res
}
