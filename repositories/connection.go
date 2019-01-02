package repositories

import (
	"github.com/jackc/pgx"
)

const (
	NotFoundErrorText = "no rows in result set"
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
			ConnConfig: c.config,
		},
	)
	return err
}

func (c *Connection) Close() error {
	c.conn.Close()
	return nil
}
