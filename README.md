# sql [![GoDoc](https://godoc.org/github.com/golangee/sql?status.svg)](http://godoc.org/github.com/golangee/sql)

A spirit of spring boot repositories for go/golang. It helps you to write *clean architecture* and *domain driven
design* styles. Define your repository just as an interface and enrich it with SQL annotations. This is
a slight violation of the *clean architecture* mantra but just keep in mind, that the interface
is your clean contract and the SQL annotation is just a part of the documentation for a concrete implementation.
There may be other annotations to allow non-sql-backends and they may co-exists with the sql annotations as
well. So nothing to feel dirty here.

## sql dialect abstraction
The notation of prepared statements is dialect specific. Even if go provides some helpers, it is in
practice not useable, because common drivers like MySQL do not support that (see [Issue 561](https://github.com/go-sql-driver/mysql/issues/561#issuecomment-337441108)).
Therefore, we bring our own named query support, which translates to the correct placeholder syntax for each
dialect (e.g. MariaDB oder PostgreSQL).

## that's so unidiomatic and java-ish
Well, it tries to keep the best of both worlds. Go is not a good fit for all use cases and we try
hard to push that border a bit further. A good article about *idiomaticity* can be found in the
[blog of Dave Cheney](https://dave.cheney.net/2020/02/23/the-zen-of-go).

## usage
This library makes heavy usage of [reflectplus](https://github.com/worldiety/reflectplus), so
be sure, that it is configured correctly. Feels best with a `go generate ./...` before launching from your
IDE. Besides that, the current implementation is reflection based, which may be changed to a generating
approach to be more performant and debugger-friendly.

### annotation example
You can only define a very limited set of methods. This is best explained by an example. Other kinds of
method specifications will not work. We do our best to validate the configuration at construction time
and give hints to solve that problem.

```go
package sms

import (
	"context"
	"github.com/golangee/uuid"
	"time"
)

type SMS struct {
	ID        uuid.UUID `ee.sql.Name:"id"`
	CreatedAt time.Time `ee.sql.Name:"created_at"`
	Text      string    `ee.sql.Name:"text"`
}

// @ee.stereotype.Repository
type Repository interface {
	// @ee.sql.Query("SELECT id,created_at,text FROM sms LIMIT :limit")
	FindAll(ctx context.Context, limit int) ([]SMS, error)

	// @ee.sql.Query("SELECT id,created_at,text FROM sms WHERE id = :id")
	FindById(ctx context.Context, id uuid.UUID) (SMS, error)

	// @ee.sql.Query("INSERT INTO sms (id,created_at,recipient,text) VALUES (:uuid, :createdAt, :recipient, :text)")
	Create(ctx context.Context, uuid uuid.UUID, createdAt time.Time, recipient string, text string) error
}
```

Don't forget to (re-)generate *reflectplus* information and proxies:
```bash
go generate ./...
```

Instantiation example
```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	sql2 "github.com/golangee/sql"
	"github.com/golangee/uuid"
	"github.com/myproject/mymodule/service/sms"
	"net/http"
	"time"
)
import _ "github.com/go-sql-driver/mysql"
import _ "github.com/myproject/mymodule" // load generated reflectplus


func main() {

	db, err := sql.Open("mysql", "user:pwd@/mydb?parseTime=true")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic("failed to connect to db: " + err.Error())
	}

	repos, err := sql2.MakeSQLRepositories(sql2.MySQL)
	if err != nil {
		panic(err)
	}
	ctx := sql2.WithContext(context.Background(), db)
	smsRepo := repos[0].(sms.Repository)
	if err := smsRepo.Create(ctx, uuid.New(), time.Now(), "1234", "hello sms"); err != nil {
		panic(err)
	}

	users, err := smsRepo.FindAll(ctx, 5)
	if err != nil {
		panic(err)
	}
	fmt.Println(users)

	usr, err := smsRepo.FindById(ctx, uuid.MustParse("04f85469-5985-48c1-91a1-bb512a71b1cf"))
	if err != nil {
		panic(err)
	}
	fmt.Println(usr)

}

```
