package repositories

import (
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	ThreadNotFoundErrMessage           = "thread not found"
	ThreadAuthorNotFoundErrMessage     = "thread author not found"
	ThreadForumNotFoundErrMessage      = "thread forum not found"
	ThreadAttributeDuplicateErrMessage = "forum attribute duplicate"
)

const (
	CreateThreadTableQuery = `
	    CREATE TABLE IF NOT EXISTS "thread" (
            "id" SERIAL
                CONSTRAINT "thread_id_pk" PRIMARY KEY,
            "slug" CITEXT
                CONSTRAINT "thread_slug_nullable" NULL
                CONSTRAINT "thread_slug_unique" UNIQUE,
            "title" TEXT
                CONSTRAINT "thread_title_not_null" NOT NULL,
            "forum" CITEXT
                CONSTRAINT "thread_forum_not_null" NOT NULL
                CONSTRAINT "thread_forum_fk" REFERENCES "forum"("slug") ON DELETE CASCADE,
            "author" CITEXT
                CONSTRAINT "thread_author_not_null" NOT NULL
                CONSTRAINT "thread_author_fk" REFERENCES "user"("nickname") ON DELETE CASCADE,
            "created_timestamp" TIMESTAMP WITH TIME ZONE
                CONSTRAINT "thread_created_timestamp_nullable" NULL,
            "message" TEXT
                CONSTRAINT "thread_message_not_null" NOT NULL,
            "num_votes" INTEGER
                DEFAULT(0)
                CONSTRAINT "thread_num_votes_not_null" NOT NULL
        );

        CREATE UNIQUE INDEX IF NOT EXISTS "thread_slug_idx" ON "thread"("slug");

        CREATE OR REPLACE FUNCTION inc_forum_num_threads()
        RETURNS TRIGGER AS
        $$
        BEGIN
            UPDATE "forum" SET
                "num_threads" = "num_threads" + 1
            WHERE "slug" = NEW."forum";
            RETURN NEW;
        END;
        $$ LANGUAGE PLPGSQL;

        DROP TRIGGER IF EXISTS "thread_insert_trg" ON "thread";

        CREATE TRIGGER "thread_insert_trg"
        AFTER INSERT ON "thread"
        FOR EACH ROW
        EXECUTE PROCEDURE inc_forum_num_threads();

        CREATE OR REPLACE FUNCTION perform_select_threads_by_forum_query(
            _forum_ TEXT, _since_ TIMESTAMPTZ, _desc_flag_ BOOLEAN, _limit_ INTEGER)
        RETURNS SETOF "thread"
        AS $$
        BEGIN
            RETURN QUERY EXECUTE FORMAT(
                'SELECT th."id",th."slug",th."title", th."forum",              ' ||
                '    th."author",th."created_timestamp",                       ' ||
                '    th."message",th."num_votes"                               ' ||
                'FROM "thread" th                                              ' ||
                'WHERE th."forum" = ''%s'' AND th."created_timestamp" > ''%s'' ' ||
                'ORDER BY th."created_timestamp"                               ' ||
                CASE
                    WHEN _desc_flag_ THEN 'DESC '
                    ELSE 'ASC '
                END ||
                CASE
                    WHEN _limit_ > 0 THEN 'LIMIT ' || _limit_::TEXT || ';'
                    ELSE ';'
                END,
            _forum_, _since_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	InsertThread          = "insert_thread"
	SelectThreadIDBySlug  = "select_thread_id_by_slug"
	SelectThreadForumByID = "select_thread_forum_by_id"
	SelectThreadByID      = "select_thread_by_id"
	SelectThreadBySlug    = "select_thread_by_slug"
	SelectThreadsByForum  = "select_threads_by_forum"

	InsertThreadQuery = `
        INSERT INTO "thread"("slug","title","forum","author","created_timestamp","message")
        VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING
        RETURNING "id";
    `
	ThreadAttributes = `
        th."id",th."slug",th."title", th."forum",th."author",
        th."created_timestamp", th."message",th."num_votes"
    `
	SelectThreadIDBySlugQuery = `
        SELECT th."id" FROM "thread" th WHERE th."slug" = $1;
    `
	SelectThreadForumByIDQuery = `
        SELECT th."forum" FROM "thread" th WHERE th."id" = $1;
    `
	SelectThreadByIDQuery = `
        SELECT ` + ThreadAttributes + `
        FROM "thread" th
        WHERE th."id" = $1;
    `
	SelectThreadBySlugQuery = `
        SELECT ` + ThreadAttributes + `
        FROM "thread" th
        WHERE th."slug" = $1;
    `
	SelectThreadsByForumQuery = `
        SELECT ` + ThreadAttributes + `
        FROM perform_select_threads_by_forum_query($1,$2,$3,$4) th;
    `
)

type ThreadRepository struct {
	conn              *Connection
	notFoundErr       *errs.Error
	conflictErr       *errs.Error
	authorNotFoundErr *errs.Error
	forumNotFoundErr  *errs.Error
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

	err = r.conn.prepareStmt(InsertThread, InsertThreadQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectThreadIDBySlug, SelectThreadIDBySlugQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectThreadForumByID, SelectThreadForumByIDQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectThreadByID, SelectThreadByIDQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectThreadBySlug, SelectThreadBySlugQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectThreadsByForum, SelectThreadsByForumQuery)
	if err != nil {
		return err
	}

	return nil
}

func (r *ThreadRepository) CreateThread(thread *models.Thread) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(SelectUserNicknameByNickname, thread.Author)
		if err := row.Scan(&thread.Author); err != nil {
			return r.authorNotFoundErr
		}

		row = tx.QueryRow(SelectForumSlugBySlug, thread.Forum)
		if err := row.Scan(&thread.Forum); err != nil {
			return r.forumNotFoundErr
		}

		slug, _ := thread.Slug.Value()
		tStp, _ := thread.CreatedTimestamp.Value()
		row = tx.QueryRow(InsertThread,
			slug, thread.Title, thread.Forum, thread.Author, tStp, thread.Message,
		)
		if err := row.Scan(&thread.ID); err == nil {
			return nil
		}

		row = tx.QueryRow(SelectThreadBySlug, slug)
		if err := r.scanThread(row.Scan, thread); err != nil {
			panic(err)
		}

		return r.conflictErr
	})
}

func (r *ThreadRepository) FindThreadIDBySlug(id *int32, slug string) *errs.Error {
	row := r.conn.conn.QueryRow(SelectThreadIDBySlug, slug)
	if err := row.Scan(id); err != nil {
		return r.notFoundErr
	}
	return nil
}

func (r *ThreadRepository) FindThreadByID(thread *models.Thread) *errs.Error {
	row := r.conn.conn.QueryRow(SelectThreadByID, thread.ID)
	if err := r.scanThread(row.Scan, thread); err != nil {
		return r.notFoundErr
	}
	return nil
}

func (r *ThreadRepository) FindThreadBySlug(thread *models.Thread) *errs.Error {
	slug, _ := thread.Slug.Value()
	row := r.conn.conn.QueryRow(SelectThreadBySlug, slug)
	if err := r.scanThread(row.Scan, thread); err != nil {
		return r.notFoundErr
	}
	return nil
}

type ForumThreadsSearchArgs struct {
	Forum string
	Since strfmt.DateTime
	Desc  bool
	Limit int
}

func (r *ThreadRepository) FindThreadsByForum(args *ForumThreadsSearchArgs) (*models.Threads, *errs.Error) {
	rows, err := r.conn.conn.Query(SelectThreadsByForum,
		args.Forum, args.Since, args.Desc, args.Limit,
	)
	if err != nil {
		return nil, r.notFoundErr
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
		return nil, r.notFoundErr
	}

	return (*models.Threads)(&threads), nil
}

type ScanFunc func(...interface{}) error

func (r *ThreadRepository) scanThread(f ScanFunc, thread *models.Thread) error {
	return f(
		&thread.ID, &thread.Slug, &thread.Title,
		&thread.Forum, &thread.Author, &thread.CreatedTimestamp,
		&thread.Message, &thread.NumVotes,
	)
}
