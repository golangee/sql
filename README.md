# sql

Package sql provides a meta-model and parsers for different sql dialect (currently only mysql).

The converter can do several things with SQL-CREATE statements:

* Multiple SQL-CREATE statements can be converted from a string to a model from the `model` package using the `parser`
  package.
* The model can be converted back into an SQL string using the `normalize` package. Attributes, quotes and properties
  are unified and sorted.
* The `diagram` package can be used to create a textual representation from a model, which can be converted into an SVG
  using the `dot` command provided by [Graphviz](https://graphviz.org/), if this is installed.

Everything can be tried out with `cmd/eesqlconv/sqlconverter.go`. With `go run cmd/eesqlconv/sqlconverter.go` the help
is displayed. For example, `go run cmd/eesqlconv/sqlconverter.go testdata/music.sql svg > test.svg` can be used to
convert the SQL from the file `music.sql` into an ER diagram.

Via `make test` all tests are started.

The grammar has already been converted into go-code, but can be generated again with `make grammar`. The
files `MySqlLexer.g4` and `MySqlParser.g4` are then translated into the folder `parser/raw`. For
this, [ANTLR](https://www.antlr.org/) must be installed.
