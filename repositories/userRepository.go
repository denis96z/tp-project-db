package repositories

import (
	"fmt"
	"github.com/jackc/pgx"
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
            "fullname" TEXT
                CONSTRAINT "user_fullname_not_null" NOT NULL,
            "email" CITEXT COLLATE "ucs_basic"
                CONSTRAINT "user_email_not_null" NOT NULL,
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

            RETURN (200, _existing_);
        END;
        $$ LANGUAGE PLPGSQL;
    `

	InsertUserStatement          = "insert_user_statement"
	SelectUserExistsByNickname   = "select_user_exists_by_nickname"
	SelectUserNicknameByNickname = "select_user_nickname_by_nickname"
	SelectUserByNickname         = "select_user_by_nickname"
	SelectUserByNicknameOrEmail  = "select_user_by_nickname_or_email"
	UpdateUserByNickname         = "update_user_by_nickname"

	UserAttributes  = `u."nickname",u."fullname",u."email",u."about"`
	InsertUserQuery = `
        DO 
    `
	SelectUserNicknameByNicknameQuery = `
        SELECT u."nickname" FROM "user" u WHERE u."nickname" = $1;
    `
	SelectUserExistsByNicknameQuery = `
        SELECT EXISTS(
            SELECT * FROM "user" WHERE "nickname" = $1
        );
    `
	SelectUserByNicknameQuery = `
        SELECT u."nickname",u."fullname",u."email",u."about"
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
        WHERE "nickname" = $1
        RETURNING "nickname","fullname","email","about";
    `
	TruncateUserTableQuery = `
        TRUNCATE "user" CASCADE;
    `
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

	err = r.conn.prepareStmt(SelectUserExistsByNickname, SelectUserExistsByNicknameQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectUserNicknameByNickname, SelectUserNicknameByNicknameQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectUserByNickname, SelectUserByNicknameQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(SelectUserByNicknameOrEmail, SelectUserByNicknameOrEmailQuery)
	if err != nil {
		return err
	}
	err = r.conn.prepareStmt(UpdateUserByNickname, UpdateUserByNicknameQuery)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) CreateUser(user *models.User, existing *string) int {
	var status int

	row := r.conn.conn.QueryRow(InsertUserStatement,
		user.Nickname, user.Email, user.FullName, user.About,
	)
	if err := row.Scan(&status, existing); err != nil {
		panic(err)
	}

	return status
}

func (r *UserRepository) FindUserByNickname(user *models.User) *errs.Error {
	row := r.conn.conn.QueryRow(SelectUserByNickname, user.Nickname)
	if err := row.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About); err != nil {
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
	query += ` ORDER BY u."nickname" `
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
		row := r.conn.conn.QueryRow(SelectForumExistsBySlugQuery, args.Forum)
		if _ = row.Scan(&exists); !exists {
			return nil, r.notFoundErr
		}
	}

	return (*models.Users)(&users), nil
}

func (r *UserRepository) UpdateUserByNickname(user *models.User) *errs.Error {
	row := r.conn.conn.QueryRow(UpdateUserByNickname,
		user.Nickname, user.FullName, user.Email, user.About,
	)
	if err := row.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About); err != nil {
		if err.Error() == NotFoundErrorText {
			return r.notFoundErr
		}
		return r.conflictErr
	}
	return nil
}

func (r *UserRepository) DeleteAllUsers() *errs.Error {
	return r.conn.performTxOp(func(tx *pgx.Tx) *errs.Error {
		_, err := tx.Exec(TruncateUserTableQuery)
		if err != nil {
			panic(err)
		}
		return nil
	})
}

func (r *UserRepository) scanUser(f ScanFunc, user *models.User) error {
	return f(
		&user.Nickname, &user.FullName, &user.Email, &user.About,
	)
}
