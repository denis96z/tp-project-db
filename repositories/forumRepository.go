package repositories

import (
	"github.com/jackc/pgx"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	ForumNotFoundErrMessage           = "forum not found"
	ForumAttributeDuplicateErrMessage = "forum attribute duplicate"
)

const (
	CreateForumTableQuery = `
	    CREATE TABLE IF NOT EXISTS "forum" (
            "slug" TEXT
                CONSTRAINT "forum_slug_pk" PRIMARY KEY,
            "title" TEXT
                CONSTRAINT "forum_title_not_null" NOT NULL,
            "admin" CITEXT
                CONSTRAINT "forum_admin_not_null" NOT NULL
                CONSTRAINT "forum_admin_fk" REFERENCES "user"("nickname") ON DELETE CASCADE,
            "num_threads" INTEGER
                DEFAULT(0)
                CONSTRAINT "forum_num_threads_not_null" NOT NULL,
            "num_posts" BIGINT
                DEFAULT(0)
                CONSTRAINT "forum_num_posts_not_null" NOT NULL
        );
    `

	InsertForum = "insert_forum"

	InsertForumQuery = `
        INSERT INTO "forum"("slug","title","admin")
        VALUES($1,$2,$3) ON CONFLICT DO NOTHING;
    `
)

type ForumRepository struct {
	conn *Connection

	insertStmt *pgx.PreparedStatement

	notFoundErr *errs.Error
	conflictErr *errs.Error
}

func NewForumRepository(conn *Connection) *ForumRepository {
	return &ForumRepository{
		conn:        conn,
		notFoundErr: errs.NewNotFoundError(ForumNotFoundErrMessage),
		conflictErr: errs.NewConflictError(ForumAttributeDuplicateErrMessage),
	}
}

func (r *ForumRepository) Init() (err error) {
	conn := r.conn.conn

	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	_, err = conn.Exec(CreateForumTableQuery)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	conn.Reset()

	tx, err = conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Commit()
		if err != nil {
			conn.Reset()
		}
	}()

	r.insertStmt, err = conn.Prepare(
		InsertForum,
		InsertForumQuery,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ForumRepository) CreateForum(forum *models.Forum) error {
	tx, err := r.conn.conn.Begin()
	if err != nil {
		panic(err)
	}
	defer func() {
		err := tx.Commit()
		if err != nil {
			panic(err)
		}
	}()

	res, err := r.conn.conn.Exec(InsertForum,
		forum.Slug, forum.Title, forum.AdminNickname,
	)
	if err != nil {
		return errs.NewInternalError(err.Error())
	}

	if res.RowsAffected() == 1 {
		return nil
	}

	rows, err := r.conn.conn.Query(SelectUserByNicknameOrEmail,
		user.Nickname, user.Email,
	)
	if err != nil {
		return errs.NewInternalError(err.Error())
	}
	defer rows.Close()

	users := make([]models.User, 0, 1)
	for rows.Next() {
		if err := rows.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About); err != nil {
			return errs.NewInternalError(err.Error())
		}
		users = append(users, *user)
	}

	*existing = models.Users(users)
	return r.conflictErr
}
