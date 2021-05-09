# tyr

Just Simple SQL Builder and DB Helper

```
$ go get -u github.com/kubuskotak/tyr
```

```go
import "github.com/kubuskotak/tyr"
```

## Driver support

* MySQL
* PostgreSQL
* SQLite3
* MsSQL

## Examples

### SelectStmt with where-value interpolation

```go
buf := NewBuffer()

ids := []int64{1, 2, 3, 4, 5}
Select("*").From("suggestions").Where("id IN ?", ids).ToSQL(dialect.MySQL, buf)
```

### SelectStmt with joins

```go
buf := NewBuffer()

Select("*").From("suggestions").
Join("subdomains", "suggestions.subdomain_id = subdomains.id").ToSQL(dialect.MySQL, buf)

Select("*").From("suggestions").
LeftJoin("subdomains", "suggestions.subdomain_id = subdomains.id").ToSQL(dialect.MySQL, buf)

// join multiple tables
Select("*").From("suggestions").
Join("subdomains", "suggestions.subdomain_id = subdomains.id").
Join("accounts", "subdomains.accounts_id = accounts.id").ToSQL(dialect.MySQL, buf)
```

### SelectStmt with raw SQL

```go
SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
```

### InsertStmt adds data from struct

```go
type Suggestion struct {
ID        int64
Title        NullString
CreatedAt    time.Time
}
sugg := &Suggestion{
Title:        NewNullString("Gopher"),
CreatedAt:    time.Now(),
}

buf := NewBuffer()
InsertInto("suggestions").
Columns("title").
Record(&sugg).ToSQL(dialect.MySQL, buf)
```

### InsertStmt adds data from value

```go
InsertInto("suggestions").
Pair("title", "Gopher").
Pair("body", "I love go.")
```

## Thanks

Inspiration and fork from these awesome libraries:

* [dbr](https://github.com/gocraft/dbr)
* [tyr](https://github.com/suryakencana007/tyr)