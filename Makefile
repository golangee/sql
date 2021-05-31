test: lint
	go test ./...

lint:
	golangci-lint run

# Build an older SQL Dialect (5.6, 5.7), which supports all relevant features
# https://github.com/antlr/grammars-v4/tree/master/sql/mysql/Positive-Technologies
# this is antlr4, e.g. brew install antlr
grammar:
	cd dialect/mysql && antlr -Dlanguage=Go -o parser -package parser MySqlLexer.g4 MySqlParser.g4
