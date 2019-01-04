package repositories

import (
	"github.com/jackc/pgx"
	"tp-project-db/consts"
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

        CREATE UNIQUE INDEX "thread_slug_idx" ON "thread"("slug");

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

        CREATE TRIGGER "thread_insert_trg"
            AFTER INSERT ON "thread"
            REFERENCING NEW TABLE AS "inserted"
            FOR EACH STATEMENT
            EXECUTE PROCEDURE inc_forum_num_threads();
    `

	InsertThread       = "insert_thread"
	SelectThreadBySlug = "select_thread_by_slug"

	InsertThreadQuery = `
        INSERT INTO "thread"("slug","title","forum","author","created_timestamp","message")
        VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING
        RETURNING "id";
    `
	SelectThreadBySlugQuery = `
        SELECT th."id",th."slug",th."title",
               th."forum",th."author",th."created_timestamp",
               th."message",th."num_votes"
        FROM "thread" th
        WHERE th."slug" = $1;
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
	err = r.conn.prepareStmt(SelectThreadBySlug, SelectThreadBySlugQuery)
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

		if thread.Forum != consts.EmptyString {
			row := tx.QueryRow(SelectForumSlugBySlug, thread.Forum)
			if err := row.Scan(&thread.Forum); err != nil {
				return r.forumNotFoundErr
			}
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
		err := row.Scan(
			&thread.ID, &thread.Slug, &thread.Title,
			&thread.Forum, &thread.Author, &thread.CreatedTimestamp,
			&thread.Message, &thread.NumVotes,
		)
		if err != nil {
			panic(err)
		}

		return r.conflictErr
	})
}
