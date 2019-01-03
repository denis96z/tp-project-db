package repositories

import "tp-project-db/errs"

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
            "forum_slug" CITEXT
                CONSTRAINT "thread_forum_slug_not_null" NOT NULL
                CONSTRAINT "thread_forum_slug_fk" REFERENCES "forum"("slug") ON DELETE CASCADE,
            "author_nickname" CITEXT
                CONSTRAINT "thread_author_nickname_not_null" NOT NULL
                CONSTRAINT "thread_author_nickname_fk" REFERENCES "user"("nickname") ON DELETE CASCADE,
            "created_timestamp" TIMESTAMP WITH TIME ZONE
                CONSTRAINT "thread_create_timestamp_not_null" NOT NULL,
            "message" TEXT
                CONSTRAINT "thread_message_not_null" NOT NULL,
            "num_votes" INTEGER
                DEFAULT(0)
                CONSTRAINT "thread_num_votes_not_null" NOT NULL
        );
    `
)

type ThreadRepository struct {
	conn             *Connection
	notFoundErr      *errs.Error
	conflictErr      *errs.Error
	adminNotFoundErr *errs.Error
}

func NewThreadRepository(conn *Connection) *ThreadRepository {
	return &ThreadRepository{
		conn:             conn,
		notFoundErr:      errs.NewNotFoundError(ThreadNotFoundErrMessage),
		conflictErr:      errs.NewConflictError(ThreadAttributeDuplicateErrMessage),
		adminNotFoundErr: errs.NewNotFoundError(AdminNotFoundErrMessage),
	}
}

func (r *ThreadRepository) Init() error {
	err := r.conn.execInit(CreateThreadTableQuery)
	if err != nil {
		return err
	}

	return nil
}
