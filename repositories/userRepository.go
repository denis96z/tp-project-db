package repositories

import (
	"database/sql"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	UserNotFoundErrMessage           = "user not found"
	UserAttributeDuplicateErrMessage = "user attribute duplicate"
)

const (
	CreateUserTableQuery = `
	    CREATE TABLE IF NOT EXISTS "user" (
            "nickname" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "user_nickname_pk" PRIMARY KEY,
            "email" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "user_email_not_null" NOT NULL,
            "fullname" TEXT
                CONSTRAINT "user_fullname_not_null" NOT NULL,
            "about" TEXT
                CONSTRAINT "user_about_not_null" NOT NULL
        );

        CREATE UNIQUE INDEX IF NOT EXISTS "user_email_idx" ON "user"("email");

        CREATE OR REPLACE FUNCTION insert_user(
             _nickname_ TEXT, _email_ TEXT, _full_name_ TEXT, _about_ TEXT
        )
        RETURNS "insert_result"
        AS $$
        DECLARE _existing_ JSON;
        BEGIN
            SELECT json_agg(json_build_object(
                u."nickname",u."email",u."fullname",u."about"
            ))
            FROM (
                SELECT u.*
                FROM "user" u
                WHERE u.nickname = _nickname_
                UNION
                SELECT u.*
                FROM "user" u
                WHERE u.email = _email_
            ) u
            INTO _existing_;

            IF _existing_ IS NOT NULL THEN
                RETURN (409, _existing_);
            END IF;

            INSERT INTO "user"("nickname","email","fullname","about")
            VALUES(_nickname_,_email_,_full_name_,_about_);

            PERFORM inc_num_users();

            RETURN (201, _existing_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	InsertUserStatement = "insert_user_statement"
)

type UserRepository struct {
	conn        *Connection
	notFoundErr *errs.Error
	conflictErr *errs.Error
}

func NewUserRepository(conn *Connection) *UserRepository {
	return &UserRepository{
		conn:        conn,
		notFoundErr: errs.NewNotFoundError(UserNotFoundErrMessage),
		conflictErr: errs.NewConflictError(UserAttributeDuplicateErrMessage),
	}
}

func (r *UserRepository) Init() error {
	err := r.conn.execInit(CreateUserTableQuery)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(InsertUserStatement, `
        SELECT * FROM insert_user($1,$2,$3,$4);
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) CreateUser(user *models.User, existing *sql.NullString) int {
	var status int

	row := r.conn.conn.QueryRow(InsertUserStatement,
		&user.Nickname, &user.Email, &user.FullName, &user.About,
	)
	if err := row.Scan(&status, existing); err != nil {
		panic(err)
	}

	return status
}
