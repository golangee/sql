package ddl

import (
	"bytes"
	"io"
)

type Generator interface {
	AsMySQL(w io.Writer) error
	AsGraphViz(w io.Writer) error
}

func ToString(f func(w io.Writer) error) (string, error) {
	b := &bytes.Buffer{}
	if err := f(b); err != nil {
		return "", err
	}

	return b.String(), nil
}

