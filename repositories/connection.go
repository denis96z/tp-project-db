package repositories

import (
	"github.com/jackc/pgx"
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

const (
	NotFoundErrorText = "no rows in result set"
)

const (
	CreateExtensionsQuery = `
        CREATE EXTENSION IF NOT EXISTS "citext";
    `
	CreateFunctionsQuery = `
        CREATE OR REPLACE FUNCTION update_value(old_value TEXT, new_value TEXT)
        RETURNS TEXT
        AS $$
            SELECT CASE
                WHEN new_value = '' THEN old_value
                ELSE new_value
            END;
        $$ LANGUAGE SQL;
    `
)

func (c *Connection) Init() error {
	_, err := c.conn.Exec(CreateExtensionsQuery)
	if err != nil {
		return err
	}
	c.conn.Reset()

	_, err = c.conn.Exec(CreateFunctionsQuery)
	if err != nil {
		return err
	}
	c.conn.Reset()

	return nil
}
