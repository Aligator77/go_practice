// Package config this part of config package create connection pools for db and configurate them
package config

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type ConnectionPool struct {
	db             *sql.DB
	DisableDBStore string
}

func NewDBConn(conf *Conf) (cp *ConnectionPool, err error) {
	cp = &ConnectionPool{
		DisableDBStore: conf.DisableDBStore,
	}
	if conf.DisableDBStore == "0" {
		connString := ""
		if conf.DB.User != "" {
			//postgres://bob:secret@1.2.3.4:5432/mydb?sslmode=verify-full
			connString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
				conf.DB.User,
				conf.DB.Password,
				conf.DB.Host,
				conf.DB.Port,
				conf.DB.Name,
			)
		}
		if conf.DB.DSN != "" {
			connString = conf.DB.DSN
		}
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
	}

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
	var d *sql.DB
	if cp.DisableDBStore == "0" {
		d = cp.db
	}
	return d
}

func (cp *ConnectionPool) Close() error {
	if cp.DisableDBStore == "0" {
		return cp.db.Close()
	}
	return nil
}

func (cp *ConnectionPool) CheckConnection(ctx context.Context) bool {
	if cp.DisableDBStore == "0" {
		err := cp.db.PingContext(ctx)
		if err != nil {
			return false
		}
	}
	return true
}
