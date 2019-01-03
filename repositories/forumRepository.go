package repositories

import (
	"github.com/jackc/pgx"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	AdminNotFoundErrMessage           = "admin not found"
	ForumNotFoundErrMessage           = "forum not found"
	ForumAttributeDuplicateErrMessage = "forum attribute duplicate"
)

const (
	CreateForumTableQuery = `
	    CREATE TABLE IF NOT EXISTS "forum" (
            "slug" CITEXT
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

	InsertForum       = "insert_forum"
	SelectForumBySlug = "select_forum_by_slug"

	InsertForumQuery = `
        INSERT INTO "forum"("slug","title","admin")
        VALUES($1,$2,$3) ON CONFLICT DO NOTHING;
    `
	SelectForumBySlugQuery = `
        SELECT f."slug",f."title",f."admin",f."num_threads",f."num_posts"
        FROM "forum" f
        WHERE f."slug" = $1;
    `
)

type ForumRepository struct {
	conn             *Connection
	notFoundErr      *errs.Error
	conflictErr      *errs.Error
	adminNotFoundErr *errs.Error
}

func NewForumRepository(conn *Connection) *ForumRepository {
	return &ForumRepository{
		conn:             conn,
		notFoundErr:      errs.NewNotFoundError(ForumNotFoundErrMessage),
		conflictErr:      errs.NewConflictError(ForumAttributeDuplicateErrMessage),
		adminNotFoundErr: errs.NewNotFoundError(AdminNotFoundErrMessage),
	}
}

func (r *ForumRepository) Init() error {
	err := r.conn.execInit(CreateForumTableQuery)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(InsertForum, InsertForumQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectForumBySlug, SelectForumBySlugQuery)
	if err != nil {
		return err
	}

	return nil
}

func (r *ForumRepository) CreateForum(forum *models.Forum) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(SelectUserNicknameByNickname, forum.AdminNickname)
		if err := row.Scan(&forum.AdminNickname); err != nil {
			return r.adminNotFoundErr
		}

		res, err := tx.Exec(InsertForum,
			forum.Slug, forum.Title, forum.AdminNickname,
		)
		if err != nil {
			panic(err)
		}
		if res.RowsAffected() == 1 {
			return nil
		}

		row = tx.QueryRow(SelectForumBySlug, forum.Slug)
		err = row.Scan(
			&forum.Slug, &forum.Title, &forum.AdminNickname, &forum.NumThreads, &forum.NumPosts,
		)
		if err != nil {
			panic(err)
		}

		return r.conflictErr
	})
}

func (r *ForumRepository) FindForumBySlug(forum *models.Forum) *errs.Error {
	row := r.conn.conn.QueryRow(SelectForumBySlug, forum.Slug)
	err := row.Scan(
		&forum.Slug, &forum.Title, &forum.AdminNickname, &forum.NumThreads, &forum.NumPosts,
	)
	if err != nil {
		return r.notFoundErr
	}
	return nil
}
