package repositories

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/jackc/pgx"
	"net/http"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	ThreadNotFoundErrMessage           = "thread not found"
	ThreadAuthorNotFoundErrMessage     = "thread author not found"
	ThreadForumNotFoundErrMessage      = "thread forum not found"
	ThreadAttributeDuplicateErrMessage = "thread attribute duplicate"
)

const (
	CreateThreadTableQuery = `
	    CREATE TABLE IF NOT EXISTS "thread" (
            "id" SERIAL
                CONSTRAINT "thread_id_pk" PRIMARY KEY,
            "slug" CITEXT
                CONSTRAINT "thread_slug_nullable" NULL,
            "title" TEXT
                CONSTRAINT "thread_title_not_null" NOT NULL,
            "forum" CITEXT
                CONSTRAINT "thread_forum_not_null" NOT NULL
                CONSTRAINT "thread_forum_fk" REFERENCES "forum"("slug"),
            "author" CITEXT
                CONSTRAINT "thread_author_not_null" NOT NULL
                CONSTRAINT "thread_author_fk" REFERENCES "user"("nickname"),
            "created_timestamp" TIMESTAMPTZ
                CONSTRAINT "thread_created_timestamp_nullable" NULL,
            "message" TEXT
                CONSTRAINT "thread_message_not_null" NOT NULL,
            "num_votes" INTEGER
                DEFAULT(0)
                CONSTRAINT "thread_num_votes_not_null" NOT NULL
        );

        CREATE INDEX IF NOT EXISTS "thread_forum_idx" ON "thread"("forum");
        CREATE INDEX IF NOT EXISTS "thread_author_idx" ON "thread"("author");
        CREATE UNIQUE INDEX IF NOT EXISTS "thread_slug_idx" ON "thread"("slug");

        CREATE OR REPLACE FUNCTION insert_thread(
            _slug_ CITEXT, _title_ TEXT, _forum_ CITEXT, _author_ CITEXT,
            _created_timestamp_ TIMESTAMPTZ, _message_ TEXT
        )
        RETURNS "query_result"
        AS $$
        DECLARE _forum_slug_ CITEXT;
        DECLARE _author_nickname_ CITEXT;
        DECLARE _existing_ JSON;
        BEGIN
            SELECT u."nickname"
            FROM "user" u
            WHERE u."nickname" = _author_
            INTO _author_nickname_;

            IF _author_nickname_ IS NULL THEN
                RETURN (404, _existing_);
            END IF;

            SELECT f."slug"
            FROM "forum" f
            WHERE f."slug" = _forum_
            INTO _forum_slug_;

            IF _forum_slug_ IS NULL THEN
                 RETURN (404, _existing_);
            END IF;

            SELECT json_build_object(
                'id', th."id", 'slug', th."slug",
                'title', th."title", 'forum', th."forum",
                'author', th."author",
                'created', th."created_timestamp",
                'message', th."message", 'votes', th."num_votes"
            )
            FROM "thread" th
            WHERE th."slug" = _slug_
            INTO _existing_;

            IF _existing_ IS NOT NULL THEN
                RETURN (409, _existing_);
            END IF;

            INSERT INTO "thread"("slug","title","forum","author","created_timestamp","message")
            VALUES(_slug_,_title_,_forum_slug_,_author_nickname_,_created_timestamp_, _message_)
            RETURNING json_build_object(
                'id', "id", 'slug', "slug",
                'title', "title", 'forum', "forum",
                'author', "author",
                'created', "created_timestamp",
                'message', "message", 'votes', "num_votes"
            ) INTO _existing_;

            UPDATE "forum" SET
                "num_threads" = "num_threads" + 1
            WHERE "slug" = _forum_slug_;

            INSERT INTO "forum_user"("forum","user")
            VALUES(_forum_slug_,_author_nickname_)
            ON CONFLICT DO NOTHING;

            RETURN (201, _existing_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	InsertThreadStatement                 = "insert_thread_statement"
	SelectThreadExistsByIDStatement       = "select_thread_exists_by_id_statement"
	SelectThreadExistsBySlugStatement     = "select_thread_exists_by_slug_statement"
	SelectThreadByIDStatement             = "select_thread_by_id_statement"
	SelectThreadBySlugStatement           = "select_thread_by_slug_statement"
	SelectThreadForumByIDStatement        = "select_thread_forum_by_id_statement"
	SelectThreadIDAndForumBySlugStatement = "select_thread_id_and_forum_by_slug_statement"
	UpdateThreadByIDStatement             = "update_thread_by_id_statement"
	UpdateThreadBySlugStatement           = "update_thread_by_slug_statement"
)

type ThreadRepository struct {
	conn              *Connection
	notFoundErr       *errs.Error
	authorNotFoundErr *errs.Error
	forumNotFoundErr  *errs.Error
	conflictErr       *errs.Error
}

func NewThreadRepository(conn *Connection) *ThreadRepository {
	return &ThreadRepository{
		conn:              conn,
		notFoundErr:       errs.NewNotFoundError(ThreadNotFoundErrMessage),
		conflictErr:       errs.NewConflictError(ThreadAttributeDuplicateErrMessage),
		authorNotFoundErr: errs.NewNotFoundError(ThreadAuthorNotFoundErrMessage),
		forumNotFoundErr:  errs.NewNotFoundError(ThreadForumNotFoundErrMessage),
	}
}

func (r *ThreadRepository) Init() error {
	err := r.conn.execInit(CreateThreadTableQuery)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(InsertThreadStatement, `
        SELECT * FROM insert_thread($1,$2,$3,$4,$5,$6);
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectThreadExistsByIDStatement, `
        SELECT EXISTS(SELECT * FROM "thread" WHERE "id" = $1);
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectThreadExistsBySlugStatement, `
        SELECT EXISTS(SELECT * FROM "thread" WHERE "slug" = $1);
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectThreadByIDStatement, `
        SELECT json_build_object(
            'id', "id", 'slug', "slug",
            'title', "title", 'forum', "forum",
            'author', "author",
            'created', "created_timestamp",
            'message', "message", 'votes', "num_votes"
        )
        FROM "thread" th
        WHERE th."id" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectThreadBySlugStatement, `
        SELECT json_build_object(
            'id', "id", 'slug', "slug",
            'title', "title", 'forum', "forum",
            'author', "author",
            'created', "created_timestamp",
            'message', "message", 'votes', "num_votes"
        )
        FROM "thread" th
        WHERE th."slug" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectThreadForumByIDStatement, `
        SELECT th."forum"
        FROM "thread" th
        WHERE th."id" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectThreadIDAndForumBySlugStatement, `
        SELECT th."id", th."forum"
        FROM "thread" th
        WHERE th."slug" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(UpdateThreadByIDStatement, `
        UPDATE "thread" SET
            ("title","message") = (
                replace_if_empty($2,"title"),
                replace_if_empty($3,"message")
            )
        WHERE "id" = $1
        RETURNING
            "id","slug","title","forum","author",
            "created_timestamp","message","num_votes";
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(UpdateThreadBySlugStatement, `
        UPDATE "thread" SET
            ("title","message") = (
                replace_if_empty($2,"title"),
                replace_if_empty($3,"message")
            )
        WHERE "slug" = $1
        RETURNING
            "id","slug","title","forum","author",
            "created_timestamp","message","num_votes";
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *ThreadRepository) CreateThread(thread *models.Thread, existing *sql.NullString) int {
	var status int

	var slug driver.Value
	if thread.Slug.Valid {
		slug = &thread.Slug.String
	} else {
		slug = nil
	}

	var createdTimestamp interface{}
	if thread.CreatedTimestamp.Valid {
		createdTimestamp = &thread.CreatedTimestamp.Timestamp
	} else {
		createdTimestamp = nil
	}

	row := r.conn.conn.QueryRow(InsertThreadStatement,
		slug, &thread.Title, &thread.Forum, &thread.Author,
		createdTimestamp, &thread.Message,
	)
	if err := row.Scan(&status, existing); err != nil {
		panic(err)
	}

	return status
}

func (r *ThreadRepository) FindThreadByID(id int32, existing *string) int {
	rows, err := r.conn.conn.Query(SelectThreadByIDStatement, &id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		err = rows.Scan(existing)
		if err != nil {
			panic(err)
		}
	}

	if !found {
		return http.StatusNotFound
	}

	return http.StatusOK
}

func (r *ThreadRepository) FindThreadBySlug(slug *string, existing *string) int {
	rows, err := r.conn.conn.Query(SelectThreadBySlugStatement, &slug)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		err = rows.Scan(existing)
		if err != nil {
			panic(err)
		}
	}

	if !found {
		return http.StatusNotFound
	}

	return http.StatusOK
}

func (r *ThreadRepository) FindThreadForumByID(args *CreatePostArgs) *errs.Error {
	row := r.conn.conn.QueryRow(SelectThreadForumByIDStatement, &args.ThreadID)
	if row.Scan(&args.ThreadForum) != nil {
		return r.notFoundErr
	}
	return nil
}

func (r *ThreadRepository) FindThreadIDAndForumBySlug(args *CreatePostArgs) *errs.Error {
	row := r.conn.conn.QueryRow(SelectThreadIDAndForumBySlugStatement, &args.ThreadSlug)
	if row.Scan(&args.ThreadID, &args.ThreadForum) != nil {
		return r.notFoundErr
	}
	return nil
}

type ForumThreadsSearchArgs struct {
	Forum string
	Since models.NullTimestamp
	Desc  bool
	Limit int
}

func (r *ThreadRepository) FindThreadsByForum(args *ForumThreadsSearchArgs) (*models.Threads, *errs.Error) {
	queryArgs := []interface{}{args.Forum}
	queryArgsCounter := 1

	query := `SELECT ` + ThreadAttributes + ` FROM "thread" th WHERE th."forum" = $1 `
	if args.Since.Valid {
		queryArgsCounter++
		queryArgs = append(queryArgs, args.Since.Timestamp)

		var eqOp string
		if args.Desc {
			eqOp = "<="
		} else {
			eqOp = ">="
		}

		query += fmt.Sprintf(`AND th."created_timestamp" %s $%d`, eqOp, queryArgsCounter)
	}
	query += ` ORDER BY th."created_timestamp"`
	if args.Desc {
		query += ` DESC`
	} else {
		query += ` ASC`
	}
	if args.Limit != 0 {
		queryArgsCounter++
		queryArgs = append(queryArgs, args.Limit)
		query += fmt.Sprintf(` LIMIT $%d;`, queryArgsCounter)
	}

	rows, err := r.conn.conn.Query(query, queryArgs...)
	if err != nil {
		return nil, r.forumNotFoundErr
	}
	defer rows.Close()

	threads := make([]models.Thread, 0)
	for rows.Next() {
		var thread models.Thread
		err = r.scanThread(rows.Scan, &thread)
		if err != nil {
			panic(err)
		}
		threads = append(threads, thread)
	}

	if len(threads) == 0 {
		var exists bool
		row := r.conn.conn.QueryRow(SelectForumExistsBySlugStatement, &args.Forum)
		if _ = row.Scan(&exists); !exists {
			return nil, r.forumNotFoundErr
		}
	}

	return (*models.Threads)(&threads), nil
}

func (r *ThreadRepository) UpdateThreadByID(thread *models.Thread) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(UpdateThreadByIDStatement,
			&thread.ID, &thread.Title, &thread.Message,
		)
		if err := r.scanThread(row.Scan, thread); err != nil {
			return r.notFoundErr
		}
		return nil
	})
}

func (r *ThreadRepository) UpdateThreadBySlug(thread *models.Thread) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(UpdateThreadBySlugStatement,
			&thread.Slug.String, &thread.Title, &thread.Message,
		)
		if err := r.scanThread(row.Scan, thread); err != nil {
			return r.notFoundErr
		}
		return nil
	})
}

func (r *ThreadRepository) scanThread(f ScanFunc, thread *models.Thread) error {
	return f(
		&thread.ID, &thread.Slug, &thread.Title,
		&thread.Forum, &thread.Author, &thread.CreatedTimestamp,
		&thread.Message, &thread.NumVotes,
	)
}
