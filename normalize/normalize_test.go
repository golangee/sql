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

package normalize_test

import (
	"fmt"
	"io/ioutil"
	"sql/dialect/mysql/parser"
	"sql/internal"
	"sql/normalize"
	"testing"
)

func TestNormalizeMusic(t *testing.T) {
	sqlBytes, _ := ioutil.ReadFile("../testdata/music.sql")
	sql := string(sqlBytes)

	// Assume that we have a correctly working parser.
	// Parse the tables, normalize them, and parse the normalized SQL again.
	// Parsing from original input should match the parsing from normalized input.
	expectedResult, err := parser.Parse(sql)
	if err != nil {
		t.Error(err)
	}

	normalized := normalize.Tables(expectedResult.Tables)

	actualResult, err := parser.Parse(normalized)
	if err != nil {
		t.Error(err)
	}

	if len(actualResult.Tables) != len(expectedResult.Tables) {
		t.Fatalf("Expected %v tables, but got %v", len(expectedResult.Tables), len(actualResult.Tables))
	}

	// Compare outputs
	for i := 0; i < len(expectedResult.Tables); i++ {
		actual := actualResult.Tables[i]
		expected := expectedResult.Tables[i]
		internal.DiffCompare(t, actual, expected, fmt.Sprintf("table %s", actual.Name))
	}
}

func TestNormalizeAlter(t *testing.T) {
	sqlBytes, _ := ioutil.ReadFile("../testdata/alter-user.sql")
	sql := string(sqlBytes)

	// Assume that we have a correctly working parser.
	// Parse the tables, normalize them, and parse the normalized SQL again.
	// Parsing from original input should match the parsing from normalized input.
	expectedResult, err := parser.Parse(sql)
	if err != nil {
		t.Error(err)
	}

	normalized := normalize.AlterStatements(expectedResult.AlterStatements)

	actualResult, err := parser.Parse(normalized)
	if err != nil {
		t.Error(err)
	}

	if len(actualResult.AlterStatements) != len(expectedResult.AlterStatements) {
		t.Fatalf("Expected %v statements, but got %v", len(expectedResult.AlterStatements), len(actualResult.AlterStatements))
	}

	// Compare outputs
	for i := 0; i < len(expectedResult.AlterStatements); i++ {
		actual := actualResult.AlterStatements[i]
		expected := expectedResult.AlterStatements[i]
		internal.DiffCompare(t, actual, expected, fmt.Sprintf("statement #%d", i))
	}
}
