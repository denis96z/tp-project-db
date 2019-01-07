package repositories

import (
	"database/sql"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	ForumNotFoundErrMessage           = "forum not found"
	ForumAdminNotFoundErrMessage      = "forum admin not found"
	ForumAttributeDuplicateErrMessage = "forum attribute duplicate"
)

const (
	CreateForumTableQuery = `
	    CREATE TABLE IF NOT EXISTS "forum" (
            "slug" CITEXT
                CONSTRAINT "forum_slug_pk" PRIMARY KEY,
            "admin" CITEXT
                CONSTRAINT "forum_admin_not_null" NOT NULL
                CONSTRAINT "forum_admin_fk" REFERENCES "user"("nickname"),
            "title" TEXT
                CONSTRAINT "forum_title_not_null" NOT NULL,
            "num_threads" INTEGER
                DEFAULT(0)
                CONSTRAINT "forum_num_threads_not_null" NOT NULL,
            "num_posts" BIGINT
                DEFAULT(0)
                CONSTRAINT "forum_num_posts_not_null" NOT NULL
        );

        CREATE INDEX IF NOT EXISTS "forum_admin_idx" ON "forum"("admin");

        CREATE OR REPLACE FUNCTION insert_forum(
            _slug_ CITEXT, _admin_ CITEXT, _title_ TEXT
        )
        RETURNS "query_result"
        AS $$
        DECLARE _existing_ JSON;
        BEGIN
            IF NOT EXISTS(SELECT * FROM "user" u WHERE u."nickname" = _admin_) THEN
                RETURN (404, _existing_);
            END IF;

            SELECT json_build_object(
                'slug', f."slug",
                'user', f."admin",
                'title', f."title",
                'threads', f."num_threads",
                'posts', f."num_posts"
            )
            FROM "forum" f
            WHERE f."slug" = _slug_
            INTO _existing_;

            IF _existing_ IS NOT NULL THEN
                RETURN (409, _existing_);
            END IF;

            INSERT INTO "forum"("slug","admin","title")
            VALUES(_slug_,_admin_,_title_)
            RETURNING json_build_object(
                'slug', "slug",
                'user', "admin",
                'title', "title",
                'threads', "num_threads",
                'posts', "num_posts"
            ) INTO _existing_;

            RETURN (201, _existing_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	InsertForumStatement = "insert_forum_statement"
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
		adminNotFoundErr: errs.NewNotFoundError(ForumAdminNotFoundErrMessage),
	}
}

func (r *ForumRepository) Init() error {
	err := r.conn.execInit(CreateForumTableQuery)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(InsertForumStatement, `
        SELECT * FROM insert_forum($1,$2,$3);
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *ForumRepository) CreateForum(forum *models.Forum, existing *sql.NullString) int {
	var status int

	row := r.conn.conn.QueryRow(InsertForumStatement,
		&forum.Slug, &forum.Admin, &forum.Title,
	)
	if err := row.Scan(&status, existing); err != nil {
		panic(err)
	}

	return status
}

/*
func (r *ForumRepository) FindForumBySlug(forum *models.Forum) *errs.Error {
	row := r.conn.conn.QueryRow(SelectForumBySlug, forum.Slug)
	err := row.Scan(
		&forum.Slug, &forum.Title, &forum.Admin, &forum.NumThreads, &forum.NumPosts,
	)
	if err != nil {
		return r.notFoundErr
	}
	return nil
}*/
