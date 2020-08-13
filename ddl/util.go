package ddl

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type strWriter struct {
	Writer io.Writer
	Err    error
}

func (w strWriter) Print(str string) {
	if w.Err != nil {
		return
	}

	_, err := w.Writer.Write([]byte(str))
	if err != nil {
		w.Err = err
	}
}

func (w strWriter) Printf(format string, args ...interface{}) {
	if w.Err != nil {
		return
	}

	_, err := w.Writer.Write([]byte(fmt.Sprintf(format, args...)))
	if err != nil {
		w.Err = err
	}
}

// snakeCaseToCamelCase converts strings like "my_snake-case" into "MySnakeCase".
func snakeCaseToCamelCase(str string) string {
	sb := &strings.Builder{}
	nextUp := true
	for _, r := range str {
		if r == '-' || r == '_' {
			nextUp = true
			continue
		}

		if nextUp {
			sb.WriteRune(unicode.ToUpper(r))
			nextUp = false
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func isSafeComment(str string) bool {
	if len(str) > 1024 {
		return false
	}

	return true
}

func isSafeName(str string) bool {
	if len(str) == 0 || len(str) > 255 {
		return false
	}

	for _, r := range str {
		if r >= 0 && r <= 9 || r >= 'a' && r <= 'z' || r == '-' || r == '_' {
			continue
		}
		return false
	}

	return true
}

func CheckPrint(w io.Writer, err *error, str string) {
	if *err != nil {
		return
	}

	if _, e := w.Write([]byte(str)); e != nil {
		*err = e
	}
}

func CheckPrintf(w io.Writer, err *error, format string, args ...interface{}) {
	if *err != nil {
		return
	}

	if _, e := w.Write([]byte(fmt.Sprintf(format, args...))); e != nil {
		*err = e
	}
}

// Check executes the given func and updates the error,
// but only if it has not been set yet. This should be
// used to handle defer and io.Closer correctly.
//
// Example
//  func do(fname string) (err error) {
//    file, err := os.Open(fname)
//    if err != nil{
//       return err
//    }
//
//    defer errors.Check(r.Close, &err)
//
//    // do stuff with file
//  }
func Check(f func() error, err *error) {
	newErr := f()

	if *err == nil {
		*err = newErr
	}
}

type StrWriter struct {
	Writer io.Writer
	Err    error
}

func (w StrWriter) Print(str string) {
	if w.Err != nil {
		return
	}

	_, err := w.Writer.Write([]byte(str))
	if err != nil {
		w.Err = err
	}
}

func (w StrWriter) Printf(format string, args ...interface{}) {
	if w.Err != nil {
		return
	}

	_, err := w.Writer.Write([]byte(fmt.Sprintf(format, args...)))
	if err != nil {
		w.Err = err
	}
}
