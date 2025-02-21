package helpers

import (
	"database/sql"
	"fmt"
)
import "context"
import "github.com/Aligator77/go_practice/internal/config"

type ConnectionPool struct {
	db *sql.DB
}

func CreateDbConn(conf *config.Conf) (cp *ConnectionPool, err error) {

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s",
		conf.DB.Host,
		conf.DB.User,
		conf.DB.Password,
		conf.DB.Port,
		conf.DB.Name,
	)

	cp.db, err = sql.Open("sqlserver", connString)
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
