package tyr

import (
	"fmt"
	"time"
)

func ExampleSelect() {
	Select("title", "body").
		From("suggestions").
		OrderBy("id").
		Limit(10)
}

func ExampleSelectStmt_Where() {
	ids := []int64{1, 2, 3, 4, 5}
	Select("*").From("suggestions").Where("id IN ?", ids)
}

func ExampleSelectStmt_Join() {
	Select("*").From("suggestions").
		Join("subdomains", "suggestions.subdomain_id = subdomains.id")

	Select("*").From("suggestions").
		LeftJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

	// join multiple tables
	Select("*").From("suggestions").
		Join("subdomains", "suggestions.subdomain_id = subdomains.id").
		Join("accounts", "subdomains.accounts_id = accounts.id")
}

func ExampleSelectBySql() {
	SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
}

func ExampleDeleteStmt() {
	DeleteFrom("suggestions").Where("id = ?", 1)
}

func ExampleAnd() {
	And(
		Or(
			Gt("created_at", "2015-09-10"),
			Lte("created_at", "2015-09-11"),
		),
		Eq("title", "hello world"),
	)
}

func ExampleI() {
	// I, identifier, can be used to quote.
	I("suggestions.id").As("id") // `suggestions`.`id`
}

func ExampleSelectStmt_As() {
	Select("count(id)").From(
		Select("*").From("suggestions").As("count"),
	)
}

func ExampleInsertStmt_Pair() {
	InsertInto("suggestions").
		Pair("title", "Gopher").
		Pair("body", "I love go.")
}

func ExampleInsertStmt_Record() {
	type Suggestion struct {
		ID        int64
		Title     NullString
		CreatedAt time.Time
	}
	sugg := &Suggestion{
		Title:     NewNullString("Gopher"),
		CreatedAt: time.Now(),
	}

	InsertInto("suggestions").
		Columns("title").
		Record(&sugg)

	// id is set automatically
	fmt.Println(sugg.ID)
}
