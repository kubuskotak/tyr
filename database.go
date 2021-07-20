package tyr

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

type Store interface {
	Notify(ctx context.Context, event Event)
	Subscriber(ctx context.Context, t EventType, fn EventFunc)
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error
}

type Driver interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRow(query string, args ...interface{}) *sql.Row
}

const (
	POSTGRES string = "postgres"
	MYSQL    string = "mysql"
)

type Sql struct {
	*sql.DB
	Event *EventHandler
}

func (s *Sql) WithTransaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	span, ctxSpan := opentracing.StartSpanFromContext(ctx, "tyr.WithTransaction")
	defer span.Finish()
	tx, err := s.DB.BeginTx(ctxSpan, nil)
	if err != nil {
		return err
	}

	if err := fn(ctxSpan, tx); err != nil {
		if errRoll := tx.Rollback(); errRoll != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, errRoll)
		}
		return err
	}
	return tx.Commit()
}

func (s *Sql) Subscriber(ctx context.Context, t EventType, fn EventFunc) {
	span, ctxSpan := opentracing.StartSpanFromContext(ctx, "tyr.Subscriber")
	defer span.Finish()
	s.Event.Handle(ctxSpan, t, fn)
}

func (s *Sql) Notify(ctx context.Context, event Event) {
	span, ctxSpan := opentracing.StartSpanFromContext(ctx, "tyr.Notify")
	defer span.Finish()
	s.Event.Event = event
	s.Event.Dispatcher(ctxSpan)
}

func (s *Sql) SetEvent(handler *EventHandler) {
	s.Event = handler
}

type SqlConnParams struct {
	Driver, Dsn string
}

func New(args SqlConnParams) (*Sql, error) {
	db, err := sql.Open(args.Driver, args.Dsn)
	if err != nil {
		panic(fmt.Errorf("cannot access your db master connection").Error())
	}

	return &Sql{DB: db}, nil
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %s: %s", e.Code, e.Message)
}

func CatchErr(err error) *Error {
	var e *Error
	// specific database error
	// postgre
	if pqErr, ok := err.(*pq.Error); ok {
		e = &Error{
			Code:    pqErr.Code.Name(),
			Message: pqErr.Message,
		}
	}
	// mysql
	if myErr, ok := err.(*mysql.MySQLError); ok {
		e = &Error{
			Code:    strconv.Itoa(int(myErr.Number)),
			Message: myErr.Message,
		}
	}
	// default
	if er, ok := err.(*Error); ok {
		e = er
	}

	return e
}
