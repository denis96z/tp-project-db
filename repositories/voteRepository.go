package repositories

import (
	"github.com/jackc/pgx"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	VoteUserNotFoundErrMessage   = "vote user not found"
	VoteThreadNotFoundErrMessage = "vote thread not found"
)

const (
	CreateVoteTableQuery = `
	    CREATE TABLE IF NOT EXISTS "vote" (
            "id" BIGSERIAL
                CONSTRAINT "vote_id_pk" PRIMARY KEY,
            "user" CITEXT
                CONSTRAINT "vote_user_not_null" NOT NULL,
            "thread" INTEGER
                CONSTRAINT "vote_thread_not_null" NOT NULL,
            "voice" TEXT
                CONSTRAINT "vote_voice_not_null" NOT NULL,
            CONSTRAINT "vote_user_thread_unique" UNIQUE("user","thread")
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

        DROP TRIGGER IF EXISTS "vote_insert_trigger";

        CREATE TRIGGER "vote_insert_trigger"
        AFTER INSERT ON "vote"
        FOR EACH ROW
        EXECUTE PROCEDURE vote_insert_trigger_func();

        CREATE OR REPLACE FUNCTION vote_update_trigger_func()
        RETURNS TRIGGER AS
        $$
        BEGIN
            UPDATE "thread" SET
                "num_votes" = "num_votes" + (NEW."voice" * (-2))
            WHERE "id" = NEW."thread";
            RETURN NEW;
        END;
        $$ LANGUAGE PLPGSQL;

        DROP TRIGGER IF EXISTS "vote_update_trigger";

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
	userNotFoundErr   *errs.Error
	threadNotFoundErr *errs.Error
}

func NewVoteRepository(conn *Connection) *VoteRepository {
	return &VoteRepository{
		conn:              conn,
		userNotFoundErr:   errs.NewNotFoundError(VoteUserNotFoundErrMessage),
		threadNotFoundErr: errs.NewConflictError(VoteThreadNotFoundErrMessage),
	}
}

func (r *VoteRepository) Init() error {
	err := r.conn.execInit(CreateVoteTableQuery)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(InsertVote, `
        INSERT INTO "vote"("user","thread","voice") VALUES($1,$2,$3);
    `)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectVoteExists, `
        SELECT EXISTS(SELECT * FROM "vote" v WHERE v."user" = $1 AND v."thread" = $2);
    `)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectVoteVoice, `
        SELECT v."voice"
        FROM "vote" v
        WHERE v."user" = $1 AND
              v."thread" = $2;
    `)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(UpdateVote, `
        UPDATE "vote" SET "voice" = $3
        WHERE "user" = $1 AND "user" = $2;
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
		row := tx.QueryRow(SelectVoteVoice, &vote.User, &vote.Thread)
		if err := row.Scan(&voice); err == nil {
			if voice == vote.Voice {
				return nil
			}
			exists = true
		}

		if !exists {
			row = tx.QueryRow(SelectUserNicknameByNickname, vote.User)
			if err := row.Scan(&vote.User); err != nil {
				return r.userNotFoundErr
			}

			row = tx.QueryRow(SelectThreadIDBySlug, vote.Thread)
			if err := row.Scan(&vote.Thread); err != nil {
				return r.threadNotFoundErr
			}

			_, err := tx.Exec(InsertVote,
				&vote.User, &vote.Thread, &vote.Voice,
			)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := tx.Exec(UpdateVote,
				&vote.User, &vote.Thread, &vote.Voice,
			)
			if err != nil {
				panic(err)
			}
		}

		return nil
	})
}
