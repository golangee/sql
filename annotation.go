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

// AnnotationQuery defines a named SQL query which is translated into dialect specific positional
// prepared statement notations. Only one annotation per Method is allowed. Each Method must have exactly
// one annotation and all named parameters must be matched to actual method parameters.
const AnnotationQuery = "ee.sql.Query"

// AnnotationRepository defines a stereotype and is currently only a marker annotation so that #MakeSQLRepositories()
// can detects repository interfaces to implement. Actually this annotation is not sql specific and may be
// shared by different repository providers (e.g. no-sql).
const AnnotationRepository = "ee.stereotype.Repository"

// AnnotationName must be used for field names to define a mapping of sql columns to fields, which is usually required
// because the notation differs (e.g. camel case vs snake case). This is SQL specific, because other repository
// providers require a different name-mapping.
const AnnotationName = "ee.sql.Name"