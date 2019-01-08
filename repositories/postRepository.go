package repositories

import (
	"database/sql"
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
	SelectPostByIDStatement                = "select_post_by_id_statement"
	SelectPostExistsByIDStatement          = "select_post_exists_by_id_statement"
	SelectPostExistsByIDAndThreadStatement = "select_post_exists_by_id_and_thread_statement"
	UpdateForumNumPostsStatement           = "update_forum_num_posts_statement"
	InsertForumUserStatement               = "insert_forum_user_statement"
	UpdatePostByIDStatement                = "update_post_by_id_statement"
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

	err = r.conn.prepareStmt(SelectPostByIDStatement, `
        SELECT `+PostAttributes+`
        FROM "post" p
        WHERE p."id" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectPostExistsByIDStatement, `
        SELECT EXISTS(SELECT * FROM "post" p WHERE p."id" = $1);
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

	err = r.conn.prepareStmt(UpdatePostByIDStatement, `
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

const (
	PostAttributes = `
        p."id",p."parent_id",p."author",
        p."forum",p."thread",p."message",
        p."created_timestamp",p."is_edited"
    `
	ThreadAttributes = `
        th."id",th."slug",th."title", th."forum",th."author",
        th."created_timestamp", th."message",th."num_votes"
    `
	ForumAttributes = `
        f."slug",f."title",f."admin",f."num_threads",f."num_posts"
    `
	UserAttributes = `u."nickname",u."fullname",u."email",u."about"`
)

func (r *PostRepository) FindPost(post *models.Post) *errs.Error {
	row := r.conn.conn.QueryRow(SelectPostByIDStatement, &post.ID)
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
			&f.Slug, &f.Title, &f.Admin, &f.NumThreads, &f.NumPosts,
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

		query += ` AND r."parent_id" = 0`

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
			row = r.conn.conn.QueryRow(SelectThreadExistsByIDStatement, &args.ThreadID.Int64)
		} else {
			row = r.conn.conn.QueryRow(SelectThreadExistsBySlugStatement, &args.ThreadSlug)
		}
		if err = row.Scan(&exists); !exists {
			return nil, r.notFoundErr
		}
	}

	return (*models.Posts)(&posts), nil
}

type ScanFunc func(...interface{}) error

func (r *PostRepository) CheckPostExists(id int64) *errs.Error {
	var exists bool
	row := r.conn.conn.QueryRow(SelectPostExistsByIDStatement, &id)
	if _ = row.Scan(&exists); !exists {
		return r.notFoundErr
	}
	return nil
}

func (r *PostRepository) UpdatePost(post *models.Post) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(UpdatePostByIDStatement, &post.ID, &post.Message)
		if err := r.scanPost(row.Scan, post); err != nil {
			return r.notFoundErr
		}
		return nil
	})
}

func (r *PostRepository) scanPost(f ScanFunc, post *models.Post) error {
	err := f(
		&post.ID, &post.ParentID, &post.Author,
		&post.Forum, &post.Thread, &post.Message,
		&post.CreatedTimestamp, &post.IsEdited,
	)
	if err != nil {
		return err
	}
	return nil
}
