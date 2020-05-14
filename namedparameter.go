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
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Dialect int

const (
	// MySQL examples
	//  WHERE col = ?
	//  VALUES(?, ?, ?)
	MySQL Dialect = 1

	// PostgreSQL examples
	//  WHERE col = $1
	//  VALUES($1, $2, $3)
	PostgreSQL Dialect = 2

	// Oracle examples
	//  WHERE col = :1
	//  VALUES(:1, :2, :3)
	Oracle Dialect = 3
)

func (d Dialect) String() string {
	switch d {
	case MySQL:
		return "mysql"
	case PostgreSQL:
		return "postgres"
	case Oracle:
		return "oracle"
	default:
		return strconv.Itoa(int(d))
	}
}

func (d Dialect) Matches(name string) bool {
	n := strings.ToLower(name)
	return n == d.String()
}

// DetectDialect tries to auto detect the dialect from the given database connection
func DetectDialect(tx DBTX) (Dialect, error) {
	rows, err := tx.QueryContext(context.Background(), "SELECT version()")
	if err != nil {
		return -1, err
	}

	defer rows.Close()
	// e.g. PostgreSQL 12.2 on x86_64-apple-darwin19.4.0, compiled by Apple clang version 11.0.3 (clang-1103.0.32.59), 64-bit
	// e.g. 10.4.11-MariaDB
	var str string
	for rows.Next() {
		if err := rows.Scan(&str); err != nil {
			return -1, err
		}
	}
	str = strings.ToLower(str)
	if strings.Contains(str, "postgresql") {
		return PostgreSQL, nil
	}

	if strings.Contains(str, "mariadb") {
		return MySQL, nil
	}

	if strings.Contains(str, "mysql") {
		return MySQL, nil
	}

	return -1, fmt.Errorf("unknown database type: %s", str)
}

var regexParamNames = regexp.MustCompile(":\\w+")

type DialectStatement struct {
	dialect   Dialect
	srcStmt   string
	statement string
	lookup    []int
}

func (s DialectStatement) String() string {
	return s.srcStmt + " as " + s.dialect.String() + " specialized to '" + s.statement + "'"
}

// Exec executes a statement filling in the arguments in the exact order as defined by prepare
func (s DialectStatement) ExecContext(db DBTX, ctx context.Context, args ...interface{}) (sql.Result, error) {
	tmp := make([]interface{}, len(s.lookup), len(s.lookup))
	for i, argIdx := range s.lookup {
		tmp[i] = args[argIdx]
	}
	return db.ExecContext(ctx, s.statement, tmp...)
}

func (s DialectStatement) QueryContext(db DBTX, ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	tmp := make([]interface{}, len(s.lookup), len(s.lookup))
	for i, argIdx := range s.lookup {
		tmp[i] = args[argIdx]
	}
	return db.QueryContext(ctx, s.statement, tmp...)
}

// A NamedParameterStatement is like a prepared statement but cross SQL dialect capable.
// Example:
//  "SELECT * FROM table WHERE x = :myParam AND y = :myParam OR z = :myOtherParam
type NamedParameterStatement string

// Validate checks if the named parameters and given names are congruent
func (s NamedParameterStatement) Validate(names []string) error {
	expectedNames := s.Names()
	if err := subset(expectedNames, names); err != nil {
		return err
	}
	return subset(names, expectedNames)
}

func (s NamedParameterStatement) Names() []string {
	names := regexParamNames.FindAllString(string(s), -1)
	for i, n := range names {
		names[i] = n[1:]
	}
	return names
}

// Prepare creates a dialect specific statement using the given argNames. Later you need to keep the exact same order.
func (s NamedParameterStatement) Prepare(sql Dialect, argNames []string) (DialectStatement, error) {
	if err := s.Validate(argNames); err != nil {
		return DialectStatement{}, err
	}

	switch sql {
	case MySQL:
		// mysql has no enumeration, so we need to repeat the according parameters
		params := s.Names()
		lookup := make([]int, len(params), len(params))
		for i, p := range params {
			for idxArg, arg := range argNames {
				if arg == p {
					lookup[i] = idxArg
					break
				}
			}
		}
		stmt := regexParamNames.ReplaceAllString(string(s), "?")
		return DialectStatement{
			dialect:   sql,
			srcStmt:   string(s),
			statement: stmt,
			lookup:    lookup,
		}, nil
	case PostgreSQL:
		fallthrough
	case Oracle:
		panic("not yet implemented")
	default:
		panic(sql)
	}
}

func subset(aSlice, bSlice []string) error {
	for _, a := range aSlice {
		found := false
		for _, b := range bSlice {
			if a == b {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("parameter '%s' is unmapped", a)
		}
	}
	return nil
}
