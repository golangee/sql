package ddl

import "strings"

// Qualifier is like follows
//  <database>.<table>.<column>
type Qualifier string

func (q Qualifier) Names() []string {
	return strings.Split(string(q), ".")
}

func (q Qualifier) First() string {
	r := q.Names()
	if len(r) > 0 {
		return r[0]
	}

	return ""
}

func (q Qualifier) Last() string {
	r := q.Names()
	if len(r) > 0 {
		return r[len(r)-1]
	}

	return ""
}

// toMysql escapes the names with backticks, so a.b.c becomes
// `a`.`b`.`c`
func (q Qualifier) toMysql() string {
	return "`" + strings.Join(q.Names(), "`.`") + "`"
}
