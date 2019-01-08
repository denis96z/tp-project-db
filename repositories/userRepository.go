package repositories

import (
	"database/sql"
	"fmt"
	"tp-project-db/consts"
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
             _nickname_ CITEXT, _email_ CITEXT, _full_name_ TEXT, _about_ TEXT
        )
        RETURNS "query_result"
        AS $$
        DECLARE _existing_ JSON;
        BEGIN
            SELECT json_agg(json_build_object(
                'nickname', u."nickname",
                'email', u."email",
                'fullname', u."fullname",
                'about', u."about"
            ))
            FROM (
                SELECT u.*
                FROM "user" u
                WHERE u."nickname" = _nickname_
                UNION
                SELECT u.*
                FROM "user" u
                WHERE u."email" = _email_
            ) u
            INTO _existing_;

            IF _existing_ IS NOT NULL THEN
                RETURN (409, _existing_);
            END IF;

            INSERT INTO "user"("nickname","email","fullname","about")
            VALUES(_nickname_,_email_,_full_name_,_about_)
            RETURNING json_build_object(
                'nickname', "nickname", 'email', "email",
                'fullname', "fullname", 'about', "about"
            ) INTO _existing_;

            RETURN (201, _existing_);
        END;
        $$ LANGUAGE PLPGSQL;

        CREATE OR REPLACE FUNCTION update_user(
             _nickname_ CITEXT, _email_ CITEXT, _full_name_ TEXT, _about_ TEXT
        )
        RETURNS "query_result"
        AS $$
        DECLARE _existing_ JSON;
        BEGIN
            SELECT json_build_object(
                'nickname', u."nickname",
                'email', u."email",
                'fullname', u."fullname",
                'about', u."about"
            )
            FROM (
                SELECT u.*
                FROM "user" u
                WHERE u."email" = _email_
            ) u
            INTO _existing_;

            IF _existing_ IS NOT NULL THEN
                RETURN (409, _existing_);
            END IF;

            UPDATE "user" SET
                "email" = CASE
                              WHEN _email_ = '' THEN "email"
                              ELSE _email_
                          END,
                "fullname" = replace_if_empty(_full_name_,"fullname"),
                "about" = replace_if_empty(_about_,"about")
            WHERE "nickname" = _nickname_
            RETURNING json_build_object(
                'nickname', "nickname", 'email', "email",
                'fullname', "fullname", 'about', "about"
            ) INTO _existing_;

            IF _existing_ IS NULL THEN
                RETURN (404, _existing_);
            END IF;

            RETURN (200, _existing_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	InsertUserStatement                   = "insert_user_statement"
	SelectUserNicknameByNicknameStatement = "select_user_nickname_by_nickname"
	SelectUserByNicknameStatement         = "select_user_by_nickname_statement"
	UpdateUserStatement                   = "update_user_statement"
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

	err = r.conn.prepareStmt(SelectUserNicknameByNicknameStatement, `
        SELECT u."nickname" FROM "user" u WHERE u."nickname" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(SelectUserByNicknameStatement, `
        SELECT u."nickname",u."email",u."fullname",u."about"
        FROM "user" u
        WHERE u."nickname" = $1;
    `)
	if err != nil {
		return err
	}

	err = r.conn.prepareStmt(UpdateUserStatement, `
        SELECT * FROM update_user($1,$2,$3,$4);
    `)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) CreateUser(user *models.User, existing *string) int {
	var status int

	row := r.conn.conn.QueryRow(InsertUserStatement,
		&user.Nickname, &user.Email, &user.FullName, &user.About,
	)
	if err := row.Scan(&status, existing); err != nil {
		panic(err)
	}

	return status
}

func (r *UserRepository) FindUser(user *models.User) *errs.Error {
	rows, err := r.conn.conn.Query(SelectUserByNicknameStatement, &user.Nickname)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		err = rows.Scan(
			&user.Nickname, &user.Email, &user.FullName, &user.About,
		)
		if err != nil {
			panic(err)
		}
	}

	if !found {
		return r.notFoundErr
	}

	return nil
}

type UsersByForumSearchArgs struct {
	Forum string
	Since string
	Desc  bool
	Limit int
}

func (r *UserRepository) FindUsersByForum(args *UsersByForumSearchArgs) (*models.Users, *errs.Error) {
	query := `
        SELECT ` + UserAttributes + `
        FROM "user" u
        JOIN "forum_user" fu ON u."nickname" = fu."user"
        WHERE fu."forum" = $1
    `
	qArgs := []interface{}{args.Forum}
	qArgsIndex := 1

	if args.Since != consts.EmptyString {
		qArgs = append(qArgs, args.Since)
		qArgsIndex++

		var eqOp string
		if args.Desc {
			eqOp = "<"
		} else {
			eqOp = ">"
		}

		query += fmt.Sprintf(`AND u."nickname" %s $%d`, eqOp, qArgsIndex)
	}
	query += ` ORDER BY lower(u."nickname") `
	if args.Desc {
		query += `DESC`
	} else {
		query += `ASC`
	}
	if args.Limit != 0 {
		qArgs = append(qArgs, args.Limit)
		qArgsIndex++
		query += fmt.Sprintf(` LIMIT $%d`, qArgsIndex)
	}
	query += `;`

	rows, err := r.conn.conn.Query(query, qArgs...)
	if err != nil {
		return nil, r.notFoundErr
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		err = r.scanUser(rows.Scan, &user)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		var exists bool
		row := r.conn.conn.QueryRow(SelectForumExistsBySlugStatement, &args.Forum)
		if _ = row.Scan(&exists); !exists {
			return nil, r.notFoundErr
		}
	}

	return (*models.Users)(&users), nil
}

func (r *UserRepository) UpdateUser(user *models.User, existing *sql.NullString) int {
	var status int

	row := r.conn.conn.QueryRow(UpdateUserStatement,
		&user.Nickname, &user.Email, &user.FullName, &user.About,
	)
	if err := row.Scan(&status, existing); err != nil {
		panic(err)
	}

	return status
}

func (r *UserRepository) scanUser(f ScanFunc, user *models.User) error {
	return f(
		&user.Nickname, &user.FullName, &user.Email, &user.About,
	)
}
