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
// one annotation per dialect and all named parameters must be matched to actual method parameters.
// Queries without dialect are candidates for all dialects, however if a specific dialect is defined, it has a higher
// priority and will not be considered a duplicate. Example:
//	@ee.sql.Query("SELECT id FROM myTable WHERE col = :param") //generic named query
//  @ee.sql.Query("dialect":"mysql", "value":"SELECT BIN_TO_UUID(id) FROM myTable WHERE col = :param") // mysql query
const AnnotationQuery = "ee.sql.Query"

// AnnotationRepository defines a stereotype and is currently only a marker annotation so that #MakeSQLRepositories()
// can detect repository interfaces to implement. Actually this annotation is not sql specific and may be
// shared by different repository providers (e.g. no-sql). Example:
//  // @ee.stereotype.Repository("myTable")
//  type MyRepository interface {
//     // @ee.sql.Query("SELECT id FROM myTable WHERE id = :id")
//     FindById(ctx context.Context, id uuid.UUID) (MyEntity, error)
//  }
const AnnotationRepository = "ee.stereotype.Repository"

// AnnotationName must be used for field names to define a mapping of sql columns to fields, which is usually required
// because the notation differs (e.g. camel case vs snake case). This is SQL specific, because other repository
// providers require a different name-mapping. Example:
//  type MyEntity struct {
//     ID uuid.UUID `ee.sql.Name:"id"`
//  }
const AnnotationName = "ee.sql.Name"

// AnnotationSchema contains generic or dialect specific statements like "CREATE TABLE" statements. They are enumerated
// by order, if nothing else has been set. Internally, the migration information is kept in the following table:
//   CREATE TABLE IF NOT EXISTS "migration_schema_history"
//   (
//    	"group"              VARCHAR(255) NOT NULL,
//    	"version"            BIGINT       NOT NULL,
//    	"script"             VARCHAR(255) NOT NULL,
//    	"type"               VARCHAR(12)  NOT NULL,
//    	"checksum"           CHAR(64)     NOT NULL,
//    	"applied_at"         TIMESTAMP    NOT NULL,
//    	"execution_duration" BIGINT       NOT NULL,
//    	"status"             VARCHAR(12)  NOT NULL,
//    	"log"                TEXT         NOT NULL,
//    	PRIMARY KEY ("group", "version")
//	 )
//
// Usage examples:
//   @ee.sql.Schema("CREATE TABLE IF NOT EXISTS `my_table` (`id` BINARY(16), PRIMARY KEY (`id`)") // generic schema
//   @ee.sql.Schema("dialect":"postgresql", "value":"CREATE TABLE `my_table` (`id` UUID, PRIMARY KEY (`id`)") // specific schema
//
// Be careful when mixing specific and non-specific schema declaration: they are just filtered and have no precedence.
// Migration order is as specified, but can be overloaded. Also each migration belongs to a group, which is by default
// the name of the stereotype.
// Complex example:
//   @ee.sql.Schema("""
//   "dialect":"mysql", "version":1, "group":"some_name", "value":
//   "CREATE TABLE IF NOT EXISTS `some_table_name`
//   (
//    	`group`              VARCHAR(255) NOT NULL,
//    	`version`            BIGINT       NOT NULL,
//    	`script`             VARCHAR(255) NOT NULL,
//    	`type`               VARCHAR(12)  NOT NULL,
//    	`checksum`           CHAR(64)     NOT NULL,
//    	`applied_at`         TIMESTAMP    NOT NULL,
//    	`execution_duration` BIGINT       NOT NULL,
//    	`status`             VARCHAR(12)  NOT NULL,
//    	`log`                TEXT         NOT NULL,
//    	PRIMARY KEY (`group`, `version`)
//	 )"
//   """)
const AnnotationSchema = "ee.sql.Schema"

// DialectValue can be used by ee.sql.Schema and ee.sql.Query
const DialectValue = "dialect"

// GroupValue can be used by ee.sql.Schema to override the default group assignment
const GroupValue = "group"

// VersionValue can be used by ee.sql.Schema to override the default version assignment (which is its applicable index)
const VersionValue = "version"