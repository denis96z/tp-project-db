package repositories

import (
	"database/sql"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	VoteAuthorNotFoundErrMessage = "vote author not found"
	VoteThreadNotFoundErrMessage = "vote thread not found"
)

const (
	CreateVoteTableQuery = `
	    CREATE TABLE IF NOT EXISTS "vote" (
            "user" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "vote_user_not_null" NOT NULL
                CONSTRAINT "vote_user_fk" REFERENCES "user"("nickname"),
            "thread" INTEGER
                CONSTRAINT "vote_thread_not_null" NOT NULL
                CONSTRAINT "vote_thread_fk" REFERENCES "thread"("id"),
            "voice" INTEGER
                CONSTRAINT "vote_voice_not_null" NOT NULL,
            CONSTRAINT "vote_user_thread_pk" PRIMARY KEY("user","thread")
        );

        CREATE OR REPLACE FUNCTION add_vote(
            _user_ CITEXT, _voice_ INTEGER,
            _thread_id_ INTEGER, _thread_slug_ CITEXT
        ) RETURNS "query_result"
        AS $$
        DECLARE _prev_ INTEGER;
        DECLARE _thread_ JSON;
        BEGIN
            IF _thread_id_ IS NULL THEN
                SELECT th."id" FROM "thread" th
                WHERE th."slug" = _thread_slug_
                INTO _thread_id_;

                IF _thread_id_ IS NULL THEN
                    RETURN (404,_thread_);
                END IF;
            ELSE
                IF NOT EXISTS (SELECT * FROM "thread" WHERE "id" = _thread_id_) THEN
                    RETURN (404,_thread_);
                END IF;
            END IF;

            IF NOT EXISTS (SELECT * FROM "user" WHERE "nickname" = _user_) THEN
                RETURN (404,_thread_);
            END IF;

            SELECT v."voice"
            FROM "vote" v
            WHERE v."user" = _user_ AND
                  v."thread" = _thread_id_
            INTO _prev_;

            IF _prev_ IS NULL THEN
                INSERT INTO "vote"("user","thread","voice")
                VALUES(_user_,_thread_id_,_voice_);

                UPDATE "thread" SET
                    "num_votes" = "num_votes" + _voice_
                WHERE "id" = _thread_id_
                RETURNING json_build_object(
                    'id', "id",'slug', "slug",'title', "title",
                    'forum', "forum",'author', "author",'created',"created_timestamp",
                    'message',"message", 'votes', "num_votes")
                INTO _thread_;
            ELSE
                IF _prev_ = _voice_ THEN
                    SELECT json_build_object(
                        'id', "id",'slug', "slug",'title', "title",
                        'forum', "forum",'author', "author",'created',"created_timestamp",
                        'message',"message", 'votes', "num_votes")
                    FROM "thread" WHERE "id" = _thread_id_
                    INTO _thread_;
                ELSE
                    UPDATE "vote" SET "voice" = _voice_
                    WHERE "user" = _user_ AND "thread" = _thread_id_;

                    UPDATE "thread" SET
                        "num_votes" = "num_votes" + (2 * _voice_)
                    WHERE "id" = _thread_id_
                    RETURNING json_build_object(
                        'id', "id",'slug', "slug",'title', "title",
                        'forum', "forum",'author', "author",'created',"created_timestamp",
                        'message',"message", 'votes', "num_votes")
                    INTO _thread_;
                END IF;
            END IF;

            RETURN (200,_thread_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	AddVoteStatement = "add_vote_statement"
)

type VoteRepository struct {
	conn              *Connection
	authorNotFoundErr *errs.Error
	threadNotFoundErr *errs.Error
}

func NewVoteRepository(conn *Connection) *VoteRepository {
	return &VoteRepository{
		conn:              conn,
		authorNotFoundErr: errs.NewNotFoundError(VoteAuthorNotFoundErrMessage),
		threadNotFoundErr: errs.NewNotFoundError(VoteThreadNotFoundErrMessage),
	}
}

func (r *VoteRepository) Init() error {
	err := r.conn.execInit(CreateVoteTableQuery)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(AddVoteStatement, `
        SELECT * FROM add_vote($1,$2,$3,$4);
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *VoteRepository) AddVote(vote *models.Vote, thread *sql.NullString) (status int) {
	var id interface{} = nil
	if vote.ThreadID != 0 {
		id = &vote.ThreadID
	}

	row := r.conn.conn.QueryRow(AddVoteStatement,
		&vote.User, &vote.Voice, id, &vote.ThreadSlug,
	)
	if err := row.Scan(&status, thread); err != nil {
		panic(err)
	}

	return status
}
