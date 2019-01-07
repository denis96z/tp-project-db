package repositories

import (
	"github.com/jackc/pgx"
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
            "id" BIGSERIAL
                CONSTRAINT "vote_id_pk" PRIMARY KEY,
            "author" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "vote_author_not_null" NOT NULL
                CONSTRAINT "vote_author_fk" REFERENCES "user"("nickname"),
            "thread" INTEGER
                CONSTRAINT "vote_thread_not_null" NOT NULL
                CONSTRAINT "vote_thread_fk" REFERENCES "thread"("id"),
            "voice" INTEGER
                CONSTRAINT "vote_voice_not_null" NOT NULL,
            CONSTRAINT "vote_author_thread_unique" UNIQUE("author","thread")
        );

        CREATE OR REPLACE FUNCTION vote_insert_trigger_func()
        RETURNS TRIGGER AS
        $$
        BEGIN
            UPDATE "thread" SET
                "num_votes" = "num_votes" + NEW."voice"
            WHERE "id" = NEW."thread";
            RETURN NEW;
        END;
        $$ LANGUAGE PLPGSQL;

        DROP TRIGGER IF EXISTS "vote_insert_trigger" ON "vote";

        CREATE TRIGGER "vote_insert_trigger"
        AFTER INSERT ON "vote"
        FOR EACH ROW
        EXECUTE PROCEDURE vote_insert_trigger_func();

        CREATE OR REPLACE FUNCTION vote_update_trigger_func()
        RETURNS TRIGGER AS
        $$
        BEGIN
            UPDATE "thread" SET
                "num_votes" =
                    CASE
                        WHEN OLD."voice" = NEW."voice" THEN "num_votes"
                        ELSE "num_votes" + (2 * NEW."voice")
                    END
            WHERE "id" = NEW."thread";
            RETURN NEW;
        END;
        $$ LANGUAGE PLPGSQL;

        DROP TRIGGER IF EXISTS "vote_update_trigger" ON "vote";

        CREATE TRIGGER "vote_update_trigger"
        AFTER UPDATE ON "vote"
        FOR EACH ROW
        EXECUTE PROCEDURE vote_update_trigger_func();
    `

	InsertVote       = "insert_vote"
	SelectVoteExists = "select_vote_exists"
	SelectVoteVoice  = "select_vote_voice"
	UpdateVote       = "update_vote"
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

	err = r.conn.prepareStmt(InsertVote, `
        INSERT INTO "vote"("author","thread","voice")
        VALUES($1,$2,$3);
    `)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectVoteExists, `
        SELECT EXISTS(SELECT * FROM "vote" v WHERE v."author" = $1 AND v."thread" = $2);
    `)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectVoteVoice, `
        SELECT v."voice"
        FROM "vote" v
        WHERE v."author" = $1 AND
              v."thread" = $2;
    `)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(UpdateVote, `
        UPDATE "vote" SET "voice" = $3
        WHERE "author" = $1 AND "thread" = $2;
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *VoteRepository) AddVote(vote *models.Vote) *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		exists := false

		var voice int32
		row := tx.QueryRow(SelectVoteVoice, &vote.Author, &vote.Thread)
		if err := row.Scan(&voice); err == nil {
			exists = true
		}

		if !exists {
			row = tx.QueryRow(SelectUserNicknameByNickname, &vote.Author)
			if err := row.Scan(&vote.Author); err != nil {
				return r.authorNotFoundErr
			}

			row = tx.QueryRow(SelectThreadExistsByIDQuery, &vote.Thread)
			if err := row.Scan(&exists); err != nil || !exists {
				return r.threadNotFoundErr
			}

			_, err := tx.Exec(InsertVote,
				&vote.Author, &vote.Thread, &vote.Voice,
			)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := tx.Exec(UpdateVote,
				&vote.Author, &vote.Thread, &vote.Voice,
			)
			if err != nil {
				panic(err)
			}
		}

		return nil
	})
}
