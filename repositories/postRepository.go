package repositories

import (
	"database/sql/driver"
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
            "id" BIGSERIAL
                CONSTRAINT "post_id_pk" PRIMARY KEY,
            "parent_id" BIGINT
                DEFAULT(0)
                CONSTRAINT "post_parent_id_nullable" NULL
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

	InsertPost           = "insert_post"
	SelectPostExistsByID = "select_post_exists_by_id"

	InsertPostQuery = `
        INSERT INTO "post"(
            "parent_id","author","forum","thread",
            "message","created_timestamp"
        )
        VALUES($1,$2,$3,$4,$5,$6)
        RETURNING "id";
    `
	SelectPostExistsByIDQuery = `
        SELECT EXISTS(SELECT * FROM "post" p WHERE p."id" = $1);
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

	err = r.conn.prepareStmt(InsertPost, InsertPostQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectPostExistsByID, SelectPostExistsByIDQuery)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) CreatePost(post *models.Post) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		if post.ParentID != 0 {
			var parentIDExists bool
			row := tx.QueryRow(SelectPostExistsByID, post.ParentID)
			if err := row.Scan(&parentIDExists); err != nil || !parentIDExists {
				return r.conflictErr
			}
		}

		row := tx.QueryRow(SelectUserNicknameByNickname, post.Author)
		if err := row.Scan(&post.Author); err != nil {
			return r.authorNotFoundErr
		}

		row = tx.QueryRow(SelectThreadForumByID, post.Thread)
		if err := row.Scan(&post.Forum); err != nil {
			return r.threadNotFoundErr
		}

		var parentID driver.Value
		if post.ParentID > 0 {
			parentID = post.ParentID
		} else {
			parentID = nil
		}

		tStp, _ := post.CreatedTimestamp.Value()

		row = tx.QueryRow(InsertPost,
			parentID, post.Author, post.Forum,
			post.Thread, post.Message, tStp,
		)
		if err := row.Scan(&post.ID); err != nil {
			panic(err)
		}

		post.IsEdited = false
		return nil
	})
}

func (r *PostRepository) scanPost(f ScanFunc, post *models.Post) error {
	return f(
		&post.ID, &post.ParentID, &post.Author,
		&post.Forum, &post.Thread, &post.Message,
		&post.CreatedTimestamp, &post.IsEdited,
	)
}
