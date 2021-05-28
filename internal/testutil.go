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

package internal

import (
	"strings"
	"testing"

	"github.com/r3labs/diff"
)

// DiffCompare compares actual and expected and fails the test with a warning should they not be equal.
// 'in' is expected to be a string that describes what things are compared, e.g. 'table User', 'Statement 5',
// so that the warning gives a better hint to the error.
func DiffCompare(t *testing.T, actual, expected interface{}, in string) {
	t.Helper()

	changes, err := diff.Diff(actual, expected)
	if err != nil {
		t.Error(err)
	} else {
		for _, change := range changes {
			t.Errorf(
				"Path '%s' in %s differs. Expected: %v, Actual: %v", strings.Join(change.Path, "."),
				in, change.From, change.To)
		}
	}
}
