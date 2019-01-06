package repositories

import (
	"fmt"
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
            "slug" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "thread_slug_nullable" NULL
                CONSTRAINT "thread_slug_unique" UNIQUE,
            "title" TEXT
                CONSTRAINT "thread_title_not_null" NOT NULL,
            "forum" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "thread_forum_not_null" NOT NULL
                CONSTRAINT "thread_forum_fk" REFERENCES "forum"("slug") ON DELETE CASCADE,
            "author" CITEXT COLLATE "ucs_basic"
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

        CREATE INDEX IF NOT EXISTS "thread_slug_idx" ON "thread"("slug");
        CREATE INDEX IF NOT EXISTS "thread_forum_idx" ON "thread"("forum");
        CREATE INDEX IF NOT EXISTS "thread_author_idx" ON "thread"("author");

        CREATE OR REPLACE FUNCTION thread_insert_trigger_func()
        RETURNS TRIGGER AS
        $$
        BEGIN
            UPDATE "forum" SET
                "num_threads" = "num_threads" + 1
            WHERE "slug" = NEW."forum";

            INSERT INTO "forum_user"("user","forum")
            VALUES(NEW."author",NEW."forum") ON CONFLICT DO NOTHING;

            RETURN NEW;
        END;
        $$ LANGUAGE PLPGSQL;

        DROP TRIGGER IF EXISTS "thread_insert_trigger" ON "thread";

        CREATE TRIGGER "thread_insert_trigger"
        AFTER INSERT ON "thread"
        FOR EACH ROW
        EXECUTE PROCEDURE thread_insert_trigger_func();
    `

	InsertThread             = "insert_thread"
	SelectThreadExistsByID   = "select_thread_exists_by_id"
	SelectThreadExistsBySlug = "select_thread_exists_by_slug"
	SelectThreadIDBySlug     = "select_thread_id_by_slug"
	SelectThreadForumByID    = "select_thread_forum_by_id"
	SelectThreadByID         = "select_thread_by_id"
	SelectThreadBySlug       = "select_thread_by_slug"
	UpdateThreadByID         = "update_thread_by_id"
	UpdateThreadBySlug       = "update_thread_by_slug"

	InsertThreadQuery = `
        INSERT INTO "thread"("slug","title","forum","author","created_timestamp","message")
        VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING
        RETURNING "id";
    `
	ThreadAttributes = `
        th."id",th."slug",th."title", th."forum",th."author",
        th."created_timestamp", th."message",th."num_votes"
    `
	SelectThreadExistsByIDQuery = `
        SELECT EXISTS(SELECT * FROM "thread" th WHERE th."id" = $1);
    `
	SelectThreadExistsBySlugQuery = `
        SELECT EXISTS(SELECT * FROM "thread" th WHERE th."slug" = $1);
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
	UpdateThreadByIDQuery = `
        UPDATE "thread" SET
            ("title","message") = (
                update_value("title",$2),
                update_value("message",$3)
            )
        WHERE "id" = $1
        RETURNING
            "id","slug","title","forum","author",
            "created_timestamp","message","num_votes";
    `
	UpdateThreadBySlugQuery = `
        UPDATE "thread" SET
            ("title","message") = (
                update_value("title",$2),
                update_value("message",$3)
            )
        WHERE "slug" = $1
        RETURNING
            "id","slug","title","forum","author",
            "created_timestamp","message","num_votes";
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
	err = r.conn.prepareStmt(SelectThreadExistsByID, SelectThreadExistsByIDQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectThreadExistsBySlug, SelectThreadExistsBySlugQuery)
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
	err = r.conn.prepareStmt(UpdateThreadByID, UpdateThreadByIDQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(UpdateThreadBySlug, UpdateThreadBySlugQuery)
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

func (r *ThreadRepository) CheckThreadExists(id int32) *errs.Error {
	var res bool
	row := r.conn.conn.QueryRow(SelectThreadExistsByID, id)
	if err := row.Scan(&res); err != nil {
		return r.notFoundErr
	}
	return nil
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
	Since models.NullTimestamp
	Desc  bool
	Limit int
}

func (r *ThreadRepository) FindThreadsByForum(args *ForumThreadsSearchArgs) (*models.Threads, *errs.Error) {
	queryArgs := []interface{}{args.Forum}
	queryArgsCounter := 1

	query := `SELECT ` + ThreadAttributes + ` FROM "thread" th WHERE th."forum" = $1 `
	if args.Since.Valid {
		queryArgsCounter++
		queryArgs = append(queryArgs, args.Since.Timestamp)

		var eqOp string
		if args.Desc {
			eqOp = "<="
		} else {
			eqOp = ">="
		}

		query += fmt.Sprintf(`AND th."created_timestamp" %s $%d`, eqOp, queryArgsCounter)
	}
	query += ` ORDER BY th."created_timestamp"`
	if args.Desc {
		query += ` DESC`
	} else {
		query += ` ASC`
	}
	if args.Limit != 0 {
		queryArgsCounter++
		queryArgs = append(queryArgs, args.Limit)
		query += fmt.Sprintf(` LIMIT $%d;`, queryArgsCounter)
	}

	rows, err := r.conn.conn.Query(query, queryArgs...)
	if err != nil {
		return nil, r.forumNotFoundErr
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
		var exists bool
		row := r.conn.conn.QueryRow(SelectForumExistsBySlugQuery, args.Forum)
		if _ = row.Scan(&exists); !exists {
			return nil, r.forumNotFoundErr
		}
	}

	return (*models.Threads)(&threads), nil
}

func (r *ThreadRepository) UpdateThreadByID(thread *models.Thread) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		row := tx.QueryRow(UpdateThreadByID,
			thread.ID, thread.Title, thread.Message,
		)
		if err := r.scanThread(row.Scan, thread); err != nil {
			return r.notFoundErr
		}
		return nil
	})
}

func (r *ThreadRepository) UpdateThreadBySlug(thread *models.Thread) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		slug, _ := thread.Slug.Value()
		row := tx.QueryRow(UpdateThreadBySlug,
			slug, thread.Title, thread.Message,
		)
		if err := r.scanThread(row.Scan, thread); err != nil {
			return r.notFoundErr
		}
		return nil
	})
}

type ScanFunc func(...interface{}) error

func (r *ThreadRepository) scanThread(f ScanFunc, thread *models.Thread) error {
	return f(
		&thread.ID, &thread.Slug, &thread.Title,
		&thread.Forum, &thread.Author, &thread.CreatedTimestamp,
		&thread.Message, &thread.NumVotes,
	)
}
