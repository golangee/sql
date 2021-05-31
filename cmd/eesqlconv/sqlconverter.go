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

package main

import (
	"flag"
	"fmt"
	"github.com/golangee/sql/ddl"
	"github.com/golangee/sql/diagram"
	"github.com/golangee/sql/dialect/mysql"
	"github.com/golangee/sql/normalize"
	"io/ioutil"
	"os"
)

const (
	OpDot       = "dot"
	OpSvg       = "svg"
	OpNormalize = "norm"
)

func main() {
	sqlFile := flag.String("sql-file", "", "the sql file to parse")
	dialect := flag.String("dialect", "mysql", "the sql dialect parser, one of (mysql)")
	operation := flag.String("op", "", "the operation to perform, one of (svg|dot|norm). 'svg' to print an svg to stdout, 'dot' to print the dot representation of the graph, 'norm' to normalize the SQL.")

	flag.Parse()

	if *sqlFile == "" || *dialect == "" || *operation == "" {
		fmt.Println("invalid usage")
		flag.PrintDefaults()
		os.Exit(-1)
		return
	}

	if err := run(*sqlFile, *dialect, *operation); err != nil {
		panic(err)
	}
}

// run actually evaluate and runs the converter command.
func run(sqlFile, dialect, op string) error {

	// Open and parse file
	fileContents, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return fmt.Errorf("cannot load sql-file '%s': %w", sqlFile, err)
	}

	var parseResult *ddl.ParseResult
	switch dialect {
	case "mysql":
		parseResult, err = mysql.Parse(string(fileContents))
		if err != nil {
			return fmt.Errorf("unable to parse mysql: %w", err)
		}
	default:
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	// Check for a valid operation.
	switch op {
	// Perform desired operation.
	case OpDot:
		dot := diagram.GenerateDot(parseResult.Tables, diagram.TwoPi)
		fmt.Println(dot)

	case OpSvg:
		dot := diagram.GenerateDot(parseResult.Tables, diagram.TwoPi)

		svg, err := diagram.DotToSvg(dot)
		if err != nil {
			return fmt.Errorf("unable to generate dot-file: %w", err)
		}

		fmt.Println(svg)

	case OpNormalize:
		normed := normalize.Tables(parseResult.Tables)
		fmt.Print(normed)
		normed = normalize.AlterStatements(parseResult.AlterStatements)
		fmt.Print(normed)
		fmt.Println()
	default:
		return fmt.Errorf("invalid operation: %s", op)
	}

	return nil
}
