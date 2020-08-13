# Architecture of *fluxexpert*
The customer *fluxexpert* is located in the wind industry and a leader of windpark management.

## Domain layers

### Domain *user*-management

#### Domain API
```go
type UUID [16]byte

type User struct {
    Id   UUID
    Name string
}

type API interface {
    // Users returns all users
    Users() ([]User, error)

    //Authenticate returns true, if the user is correct
    Authenticate(login, password string) (bool, error)
}
```

#### Persistence API
```go
type UUID [16]byte

type Group struct{
   Id        int64
   UUID      UUID
   Name      string
   Users     []*User
}

type User struct {
    Id        int64
    UUID      UUID
    Login     string
    FirstName string
    LastName  string
}

type API interface {
    SaveUser(user User) error
    DeleteUser(id int64) error
}
```

#### Persistence Implementation *mysql*

2020.08.12-16:28
```sql
CREATE TABLE IF NOT EXISTS "migration_schema_history"
(
    "group"              VARCHAR(255) NOT NULL,
    "version"            BIGINT       NOT NULL,
    "script"             VARCHAR(255) NOT NULL,
    "type"               VARCHAR(12)  NOT NULL,
    "checksum"           CHAR(64)     NOT NULL,
    "applied_at"         TIMESTAMP    NOT NULL,
    "execution_duration" BIGINT       NOT NULL,
    "status"             VARCHAR(12)  NOT NULL,
    "log"                TEXT         NOT NULL,
    PRIMARY KEY ("group", "version")
)
```

2020.08.12-16:30
```sql
ALTER TABLE "migration_schema_history"...
```

Map table *migration_schema_history* as follows
```go
import "google.com/uuid"

type MigrationSchemaHistory{
    Group string    `db:"group"`
    Version int64   `db:"version"`
    Group uuid.UUID `db:"uuid"`
}
```

Map query
```sql
SELECT `group`, `version` FROM `migration_schema` WHERE `status` = ?
```
as follows
```go
func FindById(status int64) struct{
  Group string
  Version string
}
```

Map query
```sql
SELECT `group`, `version` FROM `migration_schema` WHERE `status` = ?
```
as follows
```go
func FindById2(in struct{status int64}) (group string,version string)
```

#### Application-service 
This is a use case in Clean Architecture naming.

```go
type UUID [16]byte

type User struct {
    Id   UUID
    Name string
}

type API interface {
    // Users returns all users
    UserExists(authToken string, id UUID) (bool, error)

    // Users returns all users
    Users(authToken string)([]User,error)

    //Authenticate returns true, if the user is correct
    Authenticate(login, password string) (bool, error)
}
```

#### REST-service
TODO OpenAPI Spec?

get /api/v1/users/{id}

get /api/v1/users

post /api/v1/users/{id}/sessions

### Domain *portfolio*-management

### Domain *windpark*-management


