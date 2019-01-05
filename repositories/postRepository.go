package repositories

const (
	PostNotFoundErrMessage           = "post not found"
	PostAuthorNotFoundErrMessage     = "post author not found"
	PostForumNotFoundErrMessage      = "post forum not found"
	PostThreadNotFoundErrMessage     = "post thread not found"
	PostAttributeDuplicateErrMessage = "post attribute duplicate"
)

const (
	CreatePostTableQuery = `
        CREATE TABLE IF NOT EXISTS "post" (
            "id" BIGSERIAL
                CONSTRAINT "post_id_pk" PRIMARY KEY,
            "parent_id" BIGINT
                DEFAULT(0)
                CONSTRAINT "post_parent_id_not_null" NOT NULL
                CONSTRAINT "post_parent_id_fk" REFERENCES "post"("id") ON DELETE CASCADE,
            "author" CITEXT
                CONSTRAINT "post_author_not_null" NOT NULL
                CONSTRAINT "post_author_fk" REFERENCES "user"("nickname") ON DELETE CASCADE,
            "forum" CITEXT
                CONSTRAINT "post_forum_not_null" NOT NULL
                CONSTRAINT "post_forum_fk" REFERENCES "forum"("slug") ON DELETE CASCADE,
            "thread" INTEGER
                CONSTRAINT "post_thread_not_null" NOT NULL
                CONSTRAINT "post_thread_fk" REFERENCES "thread"("id") ON DELETE CASCADE,
            "message" TEXT
                CONSTRAINT "post_message_not_null" NOT NULL,
            "created_timestamp" TIMESTAMPTZ
                CONSTRAINT "post_created_timestamp_nullable" NULL,
            "is_edited" BOOLEAN
                DEFAULT(FALSE)
                CONSTRAINT "post_is_edited_not_null" NOT NULL
        );
    `
)
