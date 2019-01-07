package repositories

import (
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	InitServiceStatus = `
        CREATE TABLE IF NOT EXISTS "service_status"(
            "num_users" INTEGER,
            "num_forums" INTEGER,
            "num_threads" INTEGER,
            "num_posts" BIGINT
        );

        CREATE OR REPLACE FUNCTION init_status_table()
        RETURNS VOID
        AS $$
        BEGIN
            IF (SELECT COUNT(*) FROM "service_status") = 0 THEN
                INSERT INTO "service_status"(
                    "num_users","num_forums","num_threads","num_posts"
                ) VALUES(0,0,0,0);
            END IF;
        END;
        $$ LANGUAGE PLPGSQL;

        CREATE OR REPLACE FUNCTION inc_num_users()
        RETURNS VOID
        AS $$
        BEGIN
            UPDATE "service_status" SET
                "num_users" = "num_users" + 1;
            END;
        $$ LANGUAGE PLPGSQL;

        CREATE OR REPLACE FUNCTION inc_num_forums()
        RETURNS VOID
        AS $$
        BEGIN
            UPDATE "service_status" SET
                "num_forums" = "num_forums" + 1;
            END;
        $$ LANGUAGE PLPGSQL;

        CREATE OR REPLACE FUNCTION inc_num_threads(_forum_ TEXT)
        RETURNS VOID
        AS $$
        BEGIN
            UPDATE "forum" SET
                "num_threads" = "num_threads" + 1
            WHERE "slug" = _forum_;

            UPDATE "service_status" SET
                "num_threads" = "num_threads" + 1;
            END;
        $$ LANGUAGE PLPGSQL;

        CREATE OR REPLACE FUNCTION inc_num_posts(_forum_ TEXT, _diff_ BIGINT)
        RETURNS VOID
        AS $$
        BEGIN
            UPDATE "forum" SET
                "num_posts" = "num_posts" + _diff_
            WHERE "slug" = _forum_;

            UPDATE "service_status" SET
                "num_posts" = "num_posts" + _diff_;
            END;
        $$ LANGUAGE PLPGSQL;

        CREATE OR REPLACE FUNCTION clear_database()
        RETURNS VOID
        AS $$
        BEGIN
           TRUNCATE TABLE "user" CASCADE;
           UPDATE "service_status" SET (
               "num_users","num_forums","num_threads","num_posts"
           ) = (0,0,0,0);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	IncNumUsers   = "inc_num_users"
	IncNumForums  = "inc_num_forums"
	IncNumThreads = "inc_num_threads"
	IncNumPosts   = "inc_num_posts"

	SelectStatus  = "select_status"
	ClearDatabase = "clear_database"
)

type StatusRepository struct {
	conn *Connection
}

func NewStatusRepository(conn *Connection) *StatusRepository {
	return &StatusRepository{
		conn: conn,
	}
}

func (r *StatusRepository) Init() error {
	err := r.conn.execInit(InitServiceStatus)
	if err != nil {
		return err
	}

	err = r.conn.execInit(`
        DO $$ BEGIN
           PERFORM init_status_table();
        END $$;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(IncNumUsers, `
        DO $$ BEGIN
           PERFORM inc_num_users();
        END $$;
	`)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(IncNumForums, `
        DO $$ BEGIN
           PERFORM inc_num_forums();
        END $$;
	`)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(IncNumThreads, `
        DO $$ BEGIN
           PERFORM inc_num_threads($1);
        END $$;
	`)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(IncNumPosts, `
        DO $$ BEGIN
           PERFORM inc_num_posts($1,$2);
        END $$;
	`)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectStatus, `
        SELECT st."num_users", st."num_forums",
            st."num_threads", st."num_posts"
        FROM "service_status" st;
	`)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(ClearDatabase, `
        DO $$ BEGIN
           PERFORM clear_database();
        END $$;
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *StatusRepository) GetStatus(status *models.Status) *errs.Error {
	row := r.conn.conn.QueryRow(SelectStatus)
	err := row.Scan(
		&status.NumUsers, &status.NumForums,
		&status.NumThreads, &status.NumPosts,
	)
	if err != nil {
		panic(err)
	}
	return nil
}
