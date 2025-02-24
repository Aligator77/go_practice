package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"

	_ "github.com/lib/pq"

	"github.com/Aligator77/go_practice/internal/config"
)

type ConnectionPool struct {
	db *sql.DB
}

func CreateDbConn(conf *config.Conf) (cp *ConnectionPool, err error) {
	cp = new(ConnectionPool)
	//postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		conf.DB.User,
		conf.DB.Password,
		conf.DB.Host,
		conf.DB.Port,
		conf.DB.Name,
	)
	dsn, err := pq.ParseURL(connString)
	if err != nil {
		return nil, err
	}
	cp.db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = cp.db.Ping()
	if err != nil {
		cp.db.Close()
		return nil, err
	}

	cp.db.SetMaxOpenConns(conf.DB.MaxOpenCon)
	cp.db.SetMaxIdleConns(conf.DB.MaxIdleCon)

	return cp, nil
}

func (cp *ConnectionPool) Conn(ctx context.Context) (*sql.Conn, error) {
	conn, err := cp.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.PingContext(ctx)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (cp *ConnectionPool) DB() *sql.DB {
	return cp.db
}

func (cp *ConnectionPool) Close() error {
	return cp.db.Close()
}
