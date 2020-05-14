package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// The SSLMode provide a few cross platform modes, see also
//   https://dev.mysql.com/doc/refman/5.7/en/connection-options.html#option_general_ssl-mode
//   https://www.postgresql.org/docs/9.1/libpq-ssl.html
type SSLMode int

const (
	SSLPreferred      SSLMode = 0 // mysql: PREFERRED, postgres: prefer
	SSLDisable                = 1 // mysql: DISABLED, postgres: disable
	SSLRequired               = 2 // mysql: REQUIRED, postgres: require
	SSLVerifyCA               = 3 // mysql: VERIFY_CA, postgres: verify-ca
	SSLVerifyIdentify         = 4 // mysql: VERIFY_IDENTITY, postgres: verify-full

)

type Opts = struct {
	Driver       string
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSLMode      SSLMode
}

// MustOpen bails out, if it cannot connect
func MustOpen(opts Opts) *sql.DB {
	db, err := Open(opts)
	if err != nil {
		panic(err)
	}
	return db
}

// Open is a delegate to #sql.Open() and assembles the correct connection string automatically.
// You still need to import the driver, you want to support, e.g.
//    import _ "github.com/go-sql-driver/mysql" // for mysql
//    import _ "github.com/lib/pq" // for postgres
func Open(opts Opts) (*sql.DB, error) {
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	var db *sql.DB
	var err error
	switch strings.ToLower(opts.Driver) {
	case "mysql":
		if opts.Port == 0 {
			opts.Port = 3306
		}
		tls := ""
		switch opts.SSLMode {
		case SSLDisable:
			tls = "false"
		case SSLRequired:
			tls = "skip-verify"
		case SSLVerifyIdentify:
			tls = "name"
		case SSLPreferred:
			tls = "preferred"
		default:
			panic("unknown ssl mode " + strconv.Itoa(int(opts.SSLMode)))
		}
		mysqlInfo := fmt.Sprintf("%s:%s@(%s:%d)/%s?parseTime=true&tls=%s&sql_mode=ANSI", opts.User, opts.Password, opts.Host, opts.Port, opts.DatabaseName, tls)
		db, err = sql.Open("mysql", mysqlInfo)
	case "postgres":
		fallthrough
	case "postgresql":
		if opts.Port == 0 {
			opts.Port = 5432
		}
		sslmode := ""
		switch opts.SSLMode {
		case SSLDisable:
			sslmode = "disable"
		case SSLRequired:
			sslmode = "require"
		case SSLVerifyCA:
			sslmode = "verify-ca"
		case SSLVerifyIdentify:
			sslmode = "verify-full"
		case SSLPreferred:
			sslmode = "prefer"
		default:
			panic("unknown ssl mode " + strconv.Itoa(int(opts.SSLMode)))
		}
		psqlInfo := fmt.Sprintf("host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=%s",
			opts.Host, opts.Port, opts.User, opts.Password, opts.DatabaseName, sslmode)
		db, err = sql.Open("postgres", psqlInfo)
	default:
		panic("unsupported driver " + opts.Driver)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to connect %s database: %w", opts.Driver, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to ping %s database: %w", opts.Driver, err)
	}
	return db, nil
}
