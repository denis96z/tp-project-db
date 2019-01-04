package repositories

import (
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	ThreadNotFoundErrMessage           = "thread not found"
	ThreadAttributeDuplicateErrMessage = "forum attribute duplicate"
)

const (
	CreateThreadTableQuery = `
	    CREATE TABLE IF NOT EXISTS "thread" (
            "id" SERIAL
                CONSTRAINT "thread_id_pk" PRIMARY KEY,
            "slug" CITEXT
                CONSTRAINT "thread_slug_not_null" NOT NULL
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
                CONSTRAINT "thread_num_votes_not_null" NOT NULL,
            CONSTRAINT "thread_slug_forum_author_unique" UNIQUE("slug","forum","author")
        );
    `

	InsertThread = "insert_thread"

	InsertThreadQuery = `
        INSERT INTO "thread"("slug","title","forum","author","created_timestamp","message")
        VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING;
    `
)

type ThreadRepository struct {
	conn             *Connection
	notFoundErr      *errs.Error
	conflictErr      *errs.Error
	adminNotFoundErr *errs.Error
	forumNotFoundErr *errs.Error
}

func NewThreadRepository(conn *Connection) *ThreadRepository {
	return &ThreadRepository{
		conn:             conn,
		notFoundErr:      errs.NewNotFoundError(ThreadNotFoundErrMessage),
		conflictErr:      errs.NewConflictError(ThreadAttributeDuplicateErrMessage),
		adminNotFoundErr: errs.NewNotFoundError(AdminNotFoundErrMessage),
		forumNotFoundErr: errs.NewNotFoundError(ForumNotFoundErrMessage),
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

	return nil
}

func (r *ThreadRepository) CreateThread(thread *models.Thread) *errs.Error {
	return nil //TODO
}
