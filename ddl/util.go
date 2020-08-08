package ddl

import (
	"fmt"
	"github.com/golangee/reflectplus/src"
	"io"
	"strings"
	"unicode"
)

func typeDeclFromDDLKind(kind DataKind) *src.TypeDecl {
	switch kind {
	case Int64:
		return src.NewTypeDecl("int64")
	case String:
		return src.NewTypeDecl("string")
	case Timestamp:
		return src.NewTypeDecl("time.Time")
	default:
		panic("not yet implemented: " + string(kind))
	}
}

func ddlNameToGoName(str string) string {
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
