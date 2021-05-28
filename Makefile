test: lint
	go test ./...

lint:
	golangci-lint run

# Build an older SQL Dialect (5.6, 5.7), which supports all relevant features
# https://github.com/antlr/grammars-v4/tree/master/sql/mysql/Positive-Technologies
grammar:
	antlr4 -Dlanguage=Go -o dialect/mysql/parser/raw -package raw MySqlLexer.g4 MySqlParser.g4
