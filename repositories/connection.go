package repositories

import (
	"github.com/jackc/pgx"
	"tp-project-db/errs"
)

const (
	MaxConnections = 32
)

type Connection struct {
	conn   *pgx.ConnPool
	config pgx.ConnConfig
}

func NewConnection() *Connection {
	config, err := pgx.ParseEnvLibpq()
	if err != nil {
		panic(err)
	}

	return &Connection{
		config: config,
	}
}

func (c *Connection) Open() error {
	var err error
	c.conn, err = pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     c.config,
			MaxConnections: MaxConnections,
		},
	)
	return err
}

func (c *Connection) Close() error {
	c.conn.Close()
	return nil
}

const (
	CreateExtensionsQuery = `
        CREATE EXTENSION IF NOT EXISTS "citext";
    `
	CreateTypesQuery = `
        CREATE TYPE "insert_result" AS ("status" INTEGER, "result" JSON);
    `
)

func (c *Connection) Init() error {
	err := c.execInit(CreateExtensionsQuery)
	if err != nil {
		return err
	}

	err = c.execInit(CreateTypesQuery)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) execInit(stmt string) error {
	tx, err := c.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Commit()
		if err == nil {
			c.conn.Reset()
		}
	}()
	_, err = tx.Exec(stmt)
	return err
}

func (c *Connection) prepareStmt(stmt, sql string) error {
	_, err := c.conn.Prepare(stmt, sql)
	return err
}

type TxOp func(tx *pgx.Tx) *errs.Error

func (c *Connection) performTxOp(txOp TxOp) *errs.Error {
	tx, err := c.conn.Begin()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = tx.Commit()
	}()
	return txOp(tx)
}
