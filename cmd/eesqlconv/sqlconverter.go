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
	"fmt"
	"io/ioutil"
	"os"
	"sql/diagram"
	"sql/dialect/mysql/parser"
	"sql/normalize"
)

const (
	NumArgs     = 3
	OpDot       = "dot"
	OpSvg       = "svg"
	OpNormalize = "norm"
)

func main() {
	if len(os.Args) != NumArgs {
		fmt.Println("USAGE: converter <sql file> <operation>")
		fmt.Println("operation: 'svg' to print an svg to stdout")
		fmt.Println("           'dot' to print the dot representation of the graph")
		fmt.Println("           'norm' to normalize the SQL")

		return
	}

	// Open and parse file
	fileContents, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	parseResult, err := parser.Parse(string(fileContents))
	if err != nil {
		panic(err)
	}

	// Check for a valid operation.
	op := os.Args[2]
	if !isValidOp(op) {
		fmt.Println("operation must be 'svg', 'dot' or 'norm'")

		return
	}

	// Perform desired operation.
	if op == OpDot {
		dot := diagram.GenerateDot(parseResult.Tables, diagram.TwoPi)
		fmt.Println(dot)
	}

	if op == OpSvg {
		dot := diagram.GenerateDot(parseResult.Tables, diagram.TwoPi)

		svg, err := diagram.DotToSvg(dot)
		if err != nil {
			panic(err)
		}

		fmt.Println(svg)
	}

	if op == OpNormalize {
		normed := normalize.Tables(parseResult.Tables)
		fmt.Print(normed)
		normed = normalize.AlterStatements(parseResult.AlterStatements)
		fmt.Print(normed)
		fmt.Println()
	}
}

func isValidOp(op string) bool {
	return op == OpDot || op == OpSvg || op == OpNormalize
}
