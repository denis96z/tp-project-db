package repositories

import (
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
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

func (r *StatusRepository) Init() error {
	err := r.conn.prepareStmt(SelectStatus, `
        SELECT
            (SELECT COUNT(*) FROM "user")::INTEGER AS "num_users",
            (SELECT COUNT(*) FROM "forum")::INTEGER AS "num_forums",
            (SELECT COUNT(*) FROM "thread")::INTEGER AS "num_thread",
            (SELECT COUNT(*) FROM "post")::BIGINT AS "num_posts";
	`)
	if err != nil {
		return err
	}

	return nil
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
