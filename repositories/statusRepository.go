package repositories

import (
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	CreateStatusTableQuery = `
        CREATE TABLE IF NOT EXISTS "status" (
            "num_users" INTEGER DEFAULT(0),
            "num_forums" INTEGER DEFAULT(0),
            "num_threads" INTEGER DEFAULT(0),
            "num_posts" BIGINT DEFAULT(0)
        );
    `

	SelectStatus = "select_status"
)

type StatusRepository struct {
	conn *Connection
}

func NewStatusRepository(conn *Connection) *StatusRepository {
	return &StatusRepository{
		conn: conn,
	}
}

func (r *StatusRepository) GetStatus(status *models.Status) *errs.Error {
	row := r.conn.conn.QueryRow(SelectStatus)
	err := row.Scan(
		&status.NumUsers, &status.NumForums,
		&status.NumThreads, &status.NumPosts,
	)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r *StatusRepository) Init() error {
	err := r.conn.prepareStmt(SelectStatus, `
        SELECT COUNT(*) FROM "user"
        UNION
        SELECT COUNT(*) FROM "forum"
        UNION
        SELECT COUNT(*) FROM "thread"
        UNION
        SELECT COUNT(*) FROM "post";
    `)
	if err != nil {
		return err
	}

	return nil
}
