package repositories

import (
	"github.com/jackc/pgx"
	"tp-project-db/errs"
	"tp-project-db/models"
)

const (
	UserNotFoundErrMessage           = "user not found"
	UserAttributeDuplicateErrMessage = "user attribute duplicate"
)

const (
	CreateUserTableQuery = `
	    CREATE EXTENSION IF NOT EXISTS "citext";

	    CREATE TABLE "user" (
            "nickname" CITEXT
                CONSTRAINT "user_nickname_pk" PRIMARY KEY,
            "fullname" TEXT
                CONSTRAINT "user_fullname_not_null" NOT NULL,
            "email" TEXT
                CONSTRAINT "user_email_not_null" NOT NULL,
            "about" TEXT
                CONSTRAINT "user_about_not_null" NOT NULL
        );

        CREATE UNIQUE INDEX "user_nickname_email_idx" ON "user"("nickname","email");

        CREATE OR REPLACE FUNCTION update_value(old_value TEXT, new_value TEXT)
        RETURNS TEXT
        AS $$
            SELECT CASE
                WHEN new_value = '' THEN old_value
                ELSE new_value
            END;
        $$ LANGUAGE SQL;
    `

	InsertUser                  = "insert_user"
	SelectUserByNickname        = "select_user_by_nickname"
	SelectUserByNicknameOrEmail = "select_user_by_nickname_or_email"
	UpdateUserByNickname        = "update_user_by_nickname"

	InsertUserQuery = `
        INSERT INTO "user"("nickname","fullname","email","about")
        VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING;
    `

	SelectUserByNicknameQuery = `
        SELECT u."fullname",u."email",u."about"
        FROM "user" u
        WHERE u."nickname" = $1;
    `

	SelectUserByNicknameOrEmailQuery = `
        SELECT u."nickname",u."fullname",u."email",u."about"
        FROM "user" u
        WHERE u."nickname" = $1 OR u."email" = $2;
    `

	UpdateUserByNicknameQuery = `
        UPDATE "user" SET
            ("fullname","email","about") = (
                update_value("fullname",$2),
                update_value("email",$3),
                update_value("about",$4)
            )
        WHERE "nickname" = $1;
    `
)

type UserRepository struct {
	conn *Connection

	insertStmt                   *pgx.PreparedStatement
	selectByNicknameStmt         *pgx.PreparedStatement
	selectByNicknameAndEmailStmt *pgx.PreparedStatement
	updateByNicknameStmt         *pgx.PreparedStatement

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
	_, err := r.conn.conn.Exec(CreateUserTableQuery)
	if err != nil {
		return err
	}

	r.insertStmt, err = r.conn.conn.Prepare(
		InsertUser,
		InsertUserQuery,
	)
	if err != nil {
		return err
	}

	r.selectByNicknameStmt, err = r.conn.conn.Prepare(
		SelectUserByNickname,
		SelectUserByNicknameQuery,
	)
	if err != nil {
		return err
	}

	r.selectByNicknameAndEmailStmt, err = r.conn.conn.Prepare(
		SelectUserByNicknameOrEmail,
		SelectUserByNicknameOrEmailQuery,
	)
	if err != nil {
		return err
	}

	r.updateByNicknameStmt, err = r.conn.conn.Prepare(
		UpdateUserByNickname,
		UpdateUserByNicknameQuery,
	)
	if err != nil {
		return err
	}

	return err
}

func (r *UserRepository) CreateUser(user *models.User) *errs.Error {
	res, err := r.conn.conn.Exec(InsertUser,
		user.Nickname, user.FullName, user.Email, user.About,
	)
	if err != nil {
		return errs.NewInternalError(err.Error())
	}

	if res.RowsAffected() == 1 {
		return nil
	}

	row := r.conn.conn.QueryRow(SelectUserByNicknameOrEmail,
		user.Nickname, user.Email,
	)
	if err := row.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About); err != nil {
		return errs.NewInternalError(err.Error())
	}

	return r.conflictErr
}

func (r *UserRepository) FindUserByNickname(user *models.User) *errs.Error {
	row := r.conn.conn.QueryRow(SelectUserByNickname, user.Nickname)
	if err := row.Scan(&user.FullName, &user.Email, &user.About); err != nil {
		return r.notFoundErr
	}
	return nil
}

func (r *UserRepository) UpdateUserByNickname(nickname string, up *models.UserUpdate) *errs.Error {
	res, err := r.conn.conn.Exec(UpdateUserByNickname,
		up.FullName, up.Email, up.About,
	)
	if err != nil {
		return r.conflictErr
	}
	if res.RowsAffected() != 1 {
		return r.notFoundErr
	}
	return nil
}
