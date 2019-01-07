package repositories

import (
	"tp-project-db/errs"
)

const (
	PostNotFoundErrMessage       = "post not found"
	PostAuthorNotFoundErrMessage = "post author not found"
	PostForumNotFoundErrMessage  = "post forum not found"
	PostThreadNotFoundErrMessage = "post thread not found"
	PostParentNotFoundErrMessage = "post parent not found"
)

const (
	CreatePostTableQuery = `
        CREATE TABLE IF NOT EXISTS "post" (
            "id" BIGINT
                CONSTRAINT "post_id_pk" PRIMARY KEY,
            "parent_id" BIGINT
                DEFAULT(0)
                CONSTRAINT "post_parent_id_nullable" NULL,
            "author" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "post_author_not_null" NOT NULL
                CONSTRAINT "post_author_fk" REFERENCES "user"("nickname"),
            "forum" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "post_forum_not_null" NOT NULL
                CONSTRAINT "post_forum_fk" REFERENCES "forum"("slug"),
            "thread" INTEGER
                CONSTRAINT "post_thread_not_null" NOT NULL
                CONSTRAINT "post_thread_fk" REFERENCES "thread"("id"),
            "message" TEXT
                CONSTRAINT "post_message_not_null" NOT NULL,
            "created_timestamp" TIMESTAMPTZ
                CONSTRAINT "post_created_timestamp_nullable" NULL,
            "is_edited" BOOLEAN
                DEFAULT(FALSE)
                CONSTRAINT "post_is_edited_not_null" NOT NULL,
            "path" BIGINT ARRAY
        );

        CREATE SEQUENCE IF NOT EXISTS "post_id_seq" START 1;

        CREATE INDEX IF NOT EXISTS "post_author_idx" ON "post"("author");
        CREATE INDEX IF NOT EXISTS "post_forum_idx" ON "post"("forum");
        CREATE INDEX IF NOT EXISTS "post_thread_idx" ON "post"("thread");
    `
)

type PostRepository struct {
	conn              *Connection
	notFoundErr       *errs.Error
	conflictErr       *errs.Error
	authorNotFoundErr *errs.Error
	forumNotFoundErr  *errs.Error
	threadNotFoundErr *errs.Error
}

func NewPostRepository(conn *Connection) *PostRepository {
	return &PostRepository{
		conn:              conn,
		notFoundErr:       errs.NewNotFoundError(PostNotFoundErrMessage),
		conflictErr:       errs.NewConflictError(PostParentNotFoundErrMessage),
		authorNotFoundErr: errs.NewNotFoundError(PostAuthorNotFoundErrMessage),
		forumNotFoundErr:  errs.NewNotFoundError(PostForumNotFoundErrMessage),
		threadNotFoundErr: errs.NewNotFoundError(PostThreadNotFoundErrMessage),
	}
}

func (r *PostRepository) Init() error {
	err := r.conn.execInit(CreatePostTableQuery)
	if err != nil {
		return err
	}

	return nil
}
