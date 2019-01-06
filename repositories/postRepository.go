package repositories

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
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
            "author" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "post_author_not_null" NOT NULL
                CONSTRAINT "post_author_fk" REFERENCES "user"("nickname") ON DELETE CASCADE,
            "forum" CITEXT COLLATE "ucs_basic"
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
                CONSTRAINT "post_is_edited_not_null" NOT NULL,
            "path" BIGINT ARRAY
        );

        CREATE OR REPLACE FUNCTION post_insert_trigger_func()
        RETURNS TRIGGER AS
        $$
        BEGIN
            UPDATE "forum" SET
                "num_posts" = "num_posts" + 1
            WHERE "slug" = NEW."forum";

            INSERT INTO "forum_user"("user","forum")
            VALUES(NEW."author",NEW."forum") ON CONFLICT DO NOTHING;

            RETURN NEW;
        END;
        $$ LANGUAGE PLPGSQL;

        DROP TRIGGER IF EXISTS "post_insert_trigger" ON "post";

        CREATE TRIGGER "post_insert_trigger"
        AFTER INSERT ON "post"
        FOR EACH ROW
        EXECUTE PROCEDURE post_insert_trigger_func();
    `

	PostAttributes = `
        p."id",p."parent_id",p."author",
        p."forum",p."thread",p."message",
        p."created_timestamp",p."is_edited"
    `

	InsertPost                    = "insert_post"
	SelectPostByID                = "select_post_by_id"
	SelectPostExistsByIDAndThread = "select_post_exists_by_id_and_thread"
	UpdatePostByID                = "update_post_by_id"
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

	err = r.conn.prepareStmt(InsertPost, `
        INSERT INTO "post"(
            "parent_id","author","forum","thread",
            "message","created_timestamp","path"
        ) VALUES($1,$2,$3,$4,$5,$6,(
            SELECT
                CASE
                    WHEN $1::BIGINT IS NULL THEN array_append('{}', (
                        SELECT last_value FROM "post_id_seq"
                    ))
                    ELSE (
                        SELECT array_append(p."path", (
                            SELECT last_value FROM "post_id_seq"
                        ))
                        FROM "post" p
                        WHERE p."id" = $1::BIGINT
                    )
                END
        )) RETURNING "id";
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectPostByID, `
        SELECT `+PostAttributes+`
        FROM "post" p
        WHERE p."id" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectPostExistsByIDAndThread, `
        SELECT EXISTS(SELECT * FROM "post" p WHERE p."id" = $1 AND p."thread" = $2);
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(UpdatePostByID, `
        UPDATE "post" SET
            "message" = $2,
            "is_edited" =
                CASE
                    WHEN "message" != $2 THEN TRUE
                    ELSE "is_edited"
                END
        WHERE "id" = $1
        RETURNING
            "id","parent_id","author",
            "forum","thread","message",
            "created_timestamp","is_edited";
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) CreatePost(post *models.Post) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		if post.ParentID != 0 {
			var parentIDExists bool
			row := tx.QueryRow(SelectPostExistsByIDAndThread, post.ParentID, post.Thread)
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

func (r *PostRepository) FindPost(post *models.Post) *errs.Error {
	row := r.conn.conn.QueryRow(SelectPostByID, &post.ID)
	if err := r.scanPost(row.Scan, post); err != nil {
		return r.notFoundErr
	}
	return nil
}

func (r *PostRepository) FindFullPost(post *models.PostFull) *errs.Error {
	mapPtr := (*map[string]interface{})(post)

	var fAttr, fJoin string
	var thAttr, thJoin string
	var uAttr, uJoin string

	p, _ := (*mapPtr)["post"].(*models.Post)
	var pID sql.NullInt64
	dest := []interface{}{
		&p.ID, &pID, &p.Author,
		&p.Forum, &p.Thread, &p.Message,
		&p.CreatedTimestamp, &p.IsEdited,
	}

	if fItf, ok := (*mapPtr)["forum"]; ok {
		fAttr = `,` + ForumAttributes
		fJoin = ` JOIN "forum" f ON f."slug" = p."forum"`

		f := fItf.(*models.Forum)
		dest = append(dest,
			&f.Slug, &f.Title, &f.AdminNickname, &f.NumThreads, &f.NumPosts,
		)
	}
	if thItf, ok := (*mapPtr)["thread"]; ok {
		thAttr = `,` + ThreadAttributes
		thJoin = ` JOIN "thread" th ON th."id" = p."thread"`

		th := thItf.(*models.Thread)
		dest = append(dest,
			&th.ID, &th.Slug, &th.Title, &th.Forum, &th.Author,
			&th.CreatedTimestamp, &th.Message, &th.NumVotes,
		)
	}
	if uItf, ok := (*mapPtr)["author"]; ok {
		uAttr = `,` + UserAttributes
		uJoin = ` JOIN "user" u ON u."nickname" = p."author"`

		u := uItf.(*models.User)
		dest = append(dest,
			&u.Nickname, &u.FullName, &u.Email, &u.About,
		)
	}
	query := fmt.Sprintf(`SELECT %s%s%s%s FROM "post" p %s%s%s WHERE p."id" = $1;`,
		PostAttributes, fAttr, thAttr, uAttr, fJoin, thJoin, uJoin,
	)

	row := r.conn.conn.QueryRow(query, &p.ID)
	if err := row.Scan(dest...); err != nil {
		return r.notFoundErr
	}

	if pID.Valid {
		p.ParentID = pID.Int64
	} else {
		p.ParentID = 0
	}
	return nil
}

type PostsByThreadSearchArgs struct {
	ThreadID   sql.NullInt64
	ThreadSlug string
	Since      int
	SortType   string
	Desc       bool
	Limit      int
}

func (r *PostRepository) FindPostsByThread(args *PostsByThreadSearchArgs) (*models.Posts, *errs.Error) {
	query := `SELECT ` + PostAttributes + ` FROM "post" p `

	qArgs := make([]interface{}, 0, 1)
	qArgsIndex := 1

	switch args.SortType {
	case "flat":
		if !args.ThreadID.Valid {
			query += `JOIN "thread" th ON th."id" = p."thread" WHERE th."slug" = $1`
			qArgs = append(qArgs, &args.ThreadSlug)
		} else {
			query += `WHERE p."thread" = $1`
			qArgs = append(qArgs, &args.ThreadID.Int64)
		}

		if args.Since > 0 {
			qArgsIndex++
			qArgs = append(qArgs, &args.Since)

			var eqOp string
			if args.Desc {
				eqOp = "<"
			} else {
				eqOp = ">"
			}

			query += fmt.Sprintf(` AND p."id" %s $%d`, eqOp, qArgsIndex)
		}

		var sortOrd string
		if args.Desc {
			sortOrd = `DESC`
		} else {
			sortOrd = `ASC`
		}
		query += fmt.Sprintf(` ORDER BY p."id" %s`, sortOrd)

		if args.Limit > 0 {
			qArgsIndex++
			qArgs = append(qArgs, &args.Limit)
			query += fmt.Sprintf(` LIMIT $%d`, qArgsIndex)
		}

	case "tree":
		if !args.ThreadID.Valid {
			query += `JOIN "thread" th ON th."id" = p."thread" WHERE th."slug" = $1`
			qArgs = append(qArgs, &args.ThreadSlug)
		} else {
			query += `WHERE p."thread" = $1`
			qArgs = append(qArgs, &args.ThreadID.Int64)
		}

		if args.Since > 0 {
			qArgsIndex++
			qArgs = append(qArgs, &args.Since)

			var eqOp string
			if args.Desc {
				eqOp = "<"
			} else {
				eqOp = ">"
			}

			query += fmt.Sprintf(` AND p."path" %s (SELECT f."path" FROM "post" f WHERE f."id" = $%d)`, eqOp, qArgsIndex)
		}

		var sortOrd string
		if args.Desc {
			sortOrd = `DESC`
		} else {
			sortOrd = `ASC`
		}
		query += fmt.Sprintf(` ORDER BY p."path" %s`, sortOrd)

		if args.Limit > 0 {
			qArgsIndex++
			qArgs = append(qArgs, &args.Limit)
			query += fmt.Sprintf(` LIMIT $%d`, qArgsIndex)
		}

	case "parent_tree":
		query += `WHERE p."path"[1] IN (
            SELECT r."id" FROM "post" r
        `

		if !args.ThreadID.Valid {
			query += `JOIN "thread" th ON th."id" = r."thread" WHERE th."slug" = $1`
			qArgs = append(qArgs, &args.ThreadSlug)
		} else {
			query += `WHERE r."thread" = $1`
			qArgs = append(qArgs, &args.ThreadID.Int64)
		}

		query += ` AND r."parent_id" IS NULL`

		if args.Since > 0 {
			qArgsIndex++
			qArgs = append(qArgs, &args.Since)

			var eqOp string
			if args.Desc {
				eqOp = "<"
			} else {
				eqOp = ">"
			}

			query += fmt.Sprintf(` AND r."id" %s (SELECT f."path"[1] FROM "post" f WHERE f."id" = $%d)`, eqOp, qArgsIndex)
		}

		var sortOrd string
		if args.Desc {
			sortOrd = `DESC`
		} else {
			sortOrd = `ASC`
		}
		query += fmt.Sprintf(` ORDER BY r."id" %s`, sortOrd)

		if args.Limit > 0 {
			qArgsIndex++
			qArgs = append(qArgs, &args.Limit)
			query += fmt.Sprintf(` LIMIT $%d`, qArgsIndex)
		}

		query += `)`
		if args.Desc {
			query += ` ORDER BY p."path"[1] DESC, p."path"[2:]`
		} else {
			query += ` ORDER BY p."path"`
		}
	}
	query += `;`

	rows, err := r.conn.conn.Query(query, qArgs...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		var post models.Post
		if err := r.scanPost(rows.Scan, &post); err != nil {
			panic(err)
		}
		posts = append(posts, post)
	}

	if len(posts) == 0 {
		var exists bool
		var row *pgx.Row

		if args.ThreadID.Valid {
			row = r.conn.conn.QueryRow(SelectThreadExistsByID, &args.ThreadID.Int64)
		} else {
			row = r.conn.conn.QueryRow(SelectThreadExistsBySlug, &args.ThreadSlug)
		}
		if err = row.Scan(&exists); !exists {
			return nil, r.notFoundErr
		}
	}

	return (*models.Posts)(&posts), nil
}

func (r *PostRepository) UpdatePost(post *models.Post) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(UpdatePostByID, &post.ID, &post.Message)
		if err := r.scanPost(row.Scan, post); err != nil {
			return r.notFoundErr
		}
		return nil
	})
}

func (r *PostRepository) scanPost(f ScanFunc, post *models.Post) error {
	var pID sql.NullInt64
	err := f(
		&post.ID, &pID, &post.Author,
		&post.Forum, &post.Thread, &post.Message,
		&post.CreatedTimestamp, &post.IsEdited,
	)
	if err != nil {
		return err
	}

	if pID.Valid {
		post.ParentID = pID.Int64
	} else {
		post.ParentID = 0
	}
	return nil
}
