package tyr

import "github.com/kubuskotak/tyr/dialect"

func createQuery(driver string) *Query {
	var d Dialect
	switch driver {
	case "mysql":
		d = dialect.MySQL
	case "postgres", "pgx":
		d = dialect.PostgreSQL
	case "sqlite3":
		d = dialect.SQLite3
	case "mssql":
		d = dialect.MSSQL
	}
	return &Query{Dialect: d}
}

var (
	mysqlQuery    = createQuery("mysql")
	postgresQuery = createQuery("postgres")
	sqlite3Query  = createQuery("sqlite3")
	mssqlQuery    = createQuery("mssql")

	// all test Query should be here
	testQuery = []*Query{mysqlQuery, postgresQuery, sqlite3Query, mssqlQuery}
)
