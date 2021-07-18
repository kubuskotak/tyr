package tyr

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var sqlOpen = New

func TestDBConn(t *testing.T) {
	conn, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		assert.Failf(t, "failed to open stub db", "%v", err)
	}
	defer conn.Close()

	sqlOpen = func(args SqlConnParams) (*Sql, error) {
		return &Sql{Db: conn, Event: NewEventHandler()}, nil
	}

	mock.ExpectPing()

	db, cleanup, err := DBConn()

	ctx, cancel := context.WithCancel(context.Background())
	// Asserts
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, cleanup)

	db.Subscriber(ctx, UpdatedQuery, func(e Event) {
		assert.Equal(t, UpdatedQuery, e.Type)
	})

	db.Subscriber(ctx, CreatedQuery, func(e Event) {
		assert.Equal(t, CreatedQuery, e.Type)
	})

	db.Subscriber(ctx, DeletedQuery, func(e Event) {
		assert.Equal(t, DeletedQuery, e.Type)
	})

	payload := map[string]interface{}{
			"payload": "data",
		}

	db.Notify(ctx, Event{
		Type: CreatedQuery,
		Data: payload,
	})

	db.Notify(ctx, Event{
		Type: DeletedQuery,
		Data: payload,
	})

	db.Notify(ctx, Event{
		Type: UpdatedQuery,
		Data: payload,
	})

	mock.ExpectClose()
	cleanup()

	if err := mock.ExpectationsWereMet(); err != nil {
		assert.Failf(t, "there were unfulfilled expectations", "%v", err)
	}
	cancel()
}

func DBConn() (*Sql, func(), error) {
	conn, err := sqlOpen(SqlConnParams{
		Driver: "",
		Dsn:    "",
	})
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = conn.Db.Close()
	}

	conn.Db.SetConnMaxLifetime(time.Minute * time.Duration(5))
	conn.Db.SetMaxOpenConns(2)
	conn.Db.SetMaxIdleConns(2)

	if err = conn.Db.Ping(); err != nil {
		return nil, cleanup, err
	}

	return conn, cleanup, nil
}
