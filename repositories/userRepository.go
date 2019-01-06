package repositories

import (
	"fmt"
	"github.com/jackc/pgx"
	"log"
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
                CONSTRAINT "user_email_not_null" NOT NULL
                CONSTRAINT "user_email_unique" UNIQUE,
            "about" TEXT
                CONSTRAINT "user_about_not_null" NOT NULL
        );
    `

	InsertUser                   = "insert_user"
	SelectUserExistsByNickname   = "select_user_exists_by_nickname"
	SelectUserNicknameByNickname = "select_user_nickname_by_nickname"
	SelectUserByNickname         = "select_user_by_nickname"
	SelectUserByNicknameOrEmail  = "select_user_by_nickname_or_email"
	UpdateUserByNickname         = "update_user_by_nickname"

	UserAttributes  = `u."nickname",u."fullname",u."email",u."about"`
	InsertUserQuery = `
        INSERT INTO "user"("nickname","fullname","email","about")
        VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING;
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

	err = r.conn.prepareStmt(InsertUser, InsertUserQuery)
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

func (r *UserRepository) CreateUser(user *models.User, existing *models.Users) *errs.Error {
	res, err := r.conn.conn.Exec(InsertUser,
		user.Nickname, user.FullName, user.Email, user.About,
	)
	if err != nil {
		return errs.NewInternalError(err.Error())
	}

	if res.RowsAffected() == 1 {
		return nil
	}

	rows, err := r.conn.conn.Query(SelectUserByNicknameOrEmail,
		user.Nickname, user.Email,
	)
	if err != nil {
		return errs.NewInternalError(err.Error())
	}
	defer rows.Close()

	users := make([]models.User, 0, 1)
	for rows.Next() {
		if err := rows.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About); err != nil {
			return errs.NewInternalError(err.Error())
		}
		users = append(users, *user)
	}

	*existing = models.Users(users)
	return r.conflictErr
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

	log.Println(query, qArgs)

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
