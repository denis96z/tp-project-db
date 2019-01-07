package repositories

import (
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx"
	"tp-project-db/errs"
	"tp-project-db/models"
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

	SelectNextPostIDStatement              = "select_next_post_id_statement"
	SelectPostExistsByIDAndThreadStatement = "select_post_exists_by_id_and_thread_statement"
	UpdateForumNumPostsStatement           = "update_forum_num_posts_statement"
	InsertForumUserStatement               = "insert_forum_user_statement"
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

	err = r.conn.prepareStmt(SelectNextPostIDStatement, `
        SELECT nextval('post_id_seq');
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectPostExistsByIDAndThreadStatement, `
        SELECT EXISTS(SELECT * FROM "post" p WHERE p."id" = $1 AND p."thread" = $2);
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(UpdateForumNumPostsStatement, `
        UPDATE "forum" SET
            "num_posts" = "num_posts" + $2
        WHERE "slug" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(InsertForumUserStatement, `
        INSERT INTO "forum_user"("user","forum")
        VALUES($1,$2) ON CONFLICT DO NOTHING;
    `)
	if err != nil {
		return err
	}

	return nil
}

type CreatePostArgs struct {
	ThreadID    int32
	ThreadSlug  string
	ThreadForum string
	Timestamp   strfmt.DateTime
}

func (r *PostRepository) CreatePosts(posts *models.Posts, args *CreatePostArgs) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		arrPtr := (*[]models.Post)(posts)
		n := len(*arrPtr)

		query := `INSERT INTO "post"("id",
            "parent_id","author","forum","thread",
            "created_timestamp","message","path"
        )`

		qArgs := make([]interface{}, 0, n*7)
		index := 1

		for i := 0; i < n; i++ {
			postPtr := &(*arrPtr)[i]

			row := tx.QueryRow(SelectNextPostIDStatement)
			if err := row.Scan(&postPtr.ID); err != nil {
				panic(err)
			}

			postPtr.Thread = args.ThreadID
			postPtr.Forum = args.ThreadForum
			postPtr.CreatedTimestamp = args.Timestamp

			if postPtr.ParentID != 0 {
				var exists bool
				row := tx.QueryRow(SelectPostExistsByIDAndThreadStatement,
					&postPtr.ParentID, &postPtr.Thread,
				)
				if _ = row.Scan(&exists); !exists {
					return r.conflictErr
				}
			}

			row = tx.QueryRow(SelectUserNicknameByNicknameStatement, &postPtr.Author)
			if row.Scan(&postPtr.Author) != nil {
				return r.authorNotFoundErr
			}

			if i > 0 {
				query += `, `
			} else {
				query += ` VALUES`
			}
			query += fmt.Sprintf(`(%d,$%d,$%d,$%d,$%d,$%d,$%d,(
                SELECT
                    CASE
                        WHEN %d = 0 THEN ARRAY[%d]
                        ELSE (
                            SELECT array_append(p."path", %d::BIGINT)
                            FROM "post" p
                            WHERE p."id" = %d
                        )
                    END
                ))
                `,
				postPtr.ID, index, index+1, index+2, index+3, index+4, index+5,
				postPtr.ParentID, postPtr.ID, postPtr.ID, postPtr.ParentID,
			)
			index += 6
			qArgs = append(qArgs,
				&postPtr.ParentID, &postPtr.Author,
				&postPtr.Forum, &postPtr.Thread,
				&postPtr.CreatedTimestamp, &postPtr.Message,
			)
		}
		query += `;`

		res, err := tx.Exec(query, qArgs...)
		if err != nil || res.RowsAffected() != int64(n) {
			panic(err)
		}

		_, err = tx.Exec(UpdateForumNumPostsStatement, &args.ThreadForum, &n)
		if err != nil {
			panic(err)
		}

		query = `INSERT INTO "forum_user"("forum","user") VALUES`
		qArgs = make([]interface{}, 0, 2*n)
		index = 1

		for i := 0; i < n; i++ {
			if i > 0 {
				query += `, `
			}

			query += fmt.Sprintf(`($%d,$%d)`, index, index+1)
			index += 2
			qArgs = append(qArgs, &args.ThreadForum, &(*arrPtr)[i].Author)
		}

		query += ` ON CONFLICT DO NOTHING;`

		_, err = tx.Exec(query, qArgs...)
		if err != nil {
			panic(err)
		}

		return nil
	})
}
