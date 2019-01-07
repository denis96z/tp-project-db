package repositories

import (
	"tp-project-db/models"
)

const (
	InitServiceStatus = `
        CREATE OR REPLACE FUNCTION clear_database()
        RETURNS VOID
        AS $$
        BEGIN
           TRUNCATE TABLE "user" CASCADE;
           UPDATE "service_status" SET (
               "num_users","num_forums","num_threads","num_posts"
           ) = (0,0,0,0);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	SelectStatus  = "select_status"
	ClearDatabase = "clear_database"
)

type StatusRepository struct {
	conn *Connection
}

func NewStatusRepository(conn *Connection) *StatusRepository {
	return &StatusRepository{
		conn: conn,
	}
}

func (r *StatusRepository) Init() error {
	err := r.conn.execInit(InitServiceStatus)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectStatus, `
        SELECT
            (SELECT COUNT(*) FROM "user" u) AS "num_users",
            0 AS "num_forums",
            0 AS "num_threads",
            0::BIGINT AS "num_posts";
	`)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(ClearDatabase, `
        DO $$ BEGIN
           PERFORM clear_database();
        END $$;
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *StatusRepository) GetStatus(status *models.Status) {
	row := r.conn.conn.QueryRow(SelectStatus)
	err := row.Scan(
		&status.NumUsers, &status.NumForums,
		&status.NumThreads, &status.NumPosts,
	)
	if err != nil {
		panic(err)
	}
}

func (r *StatusRepository) ClearDatabase() {
	if _, err := r.conn.conn.Exec(ClearDatabase); err != nil {
		panic(err)
	}
}
