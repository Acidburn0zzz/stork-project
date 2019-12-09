package dbmodel

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/go-pg/pg/v9"
	"isc.org/stork/server/database"
)

// Represents a user held in system_user table in the database.
type SystemUser struct {
	Id           int
	Login        string
	Email        string
	Lastname     string
	Name         string
	Password     string `pg:"password_hash"`
}

type SystemUsers []SystemUser

type SystemUserOrderBy int

const (
	SystemUserOrderById SystemUserOrderBy = iota
	SystemUserOrderByLoginEmail
)

// Returns user's identity for logging purposes. It includes login, email or both.
func (u *SystemUser) Identity() string {
	// Include login, if present.
	var s string
	if len(u.Login) > 0 {
		s = fmt.Sprintf("login=%s", u.Login)
	}

	// Include email if present.
	if len(u.Email) > 0 {
		if len(s) > 0 {
			s += " "
		}
		s += fmt.Sprintf("email=%s", u.Email)
	}

	// Neither login nor email set.
	if len(s) == 0 {
		s = "unknown"
	}

	return s
}

// Inserts or updates user information in the database. If user id is set, the user information
// is updated in the database. Otherwise, the new user will be inserted. The returned
// conflict value indicates if there was a conflict while inserting/updating the user
// in the database, i.e. login/email already exists, or the updated user doesn't exist.
func (u *SystemUser) Persist(db *pg.DB) (err error, conflict bool) {
	if u.Id == 0 {
		// Add new user as the id is not set.
		err = db.Insert(u)

	} else {
		// Update existing user by primary key.
		err = db.Update(u)
		if err == pg.ErrNoRows {
			conflict = true
		}
	}
	if err != nil {
		pgErr, ok := err.(pg.Error); if ok {
			conflict = pgErr.IntegrityViolation()
		}
		errors.Wrapf(err, "database operation error while trying to persist user %s", u.Identity())
	}
	return err, conflict
}

// Sets new password for the given user id.
func SetPassword(db *pg.DB, id int, password string) (err error) {
	user := SystemUser{
		Id: id,
		Password: password,
	}

	result, err := db.Model(&user).Column("password_hash").WherePK().Update()
	if err != nil {
		errors.Wrapf(err, "database operation error while trying to set new password for the user id %d",
			id)

	} else if result.RowsAffected() == 0 {
		err = errors.Errorf("failed to update password for non existing user with id %d", id)
	}

	return err
}

// Checks if the old password matches the one in the database and modifies
// it to the new password if it does.
func ChangePassword(db *pg.DB, id int, oldPassword, newPassword string) (bool, error) {
	user := SystemUser{
		Id: id,
	}
	ok, err := db.Model(&user).
		Where("password_hash = crypt(?, password_hash) AND (id = ?)",
		oldPassword, id).Exists()

	if err != nil {
		errors.Wrapf(err, "database operation error while trying to change password of user with id %d", id)
		return false, err
	}

	if !ok {
		return false, nil
	}

	err = SetPassword(db, id, newPassword)
	return true, err
}

// Finds the user in the database by login or email and verifies that the provided password
// is correct.
func Authenticate(db *pg.DB, user *SystemUser) (bool, error) {
	// Using authentication technique described here: https://www.postgresql.org/docs/8.3/pgcrypto.html
	err := db.Model(user).
		Where("password_hash = crypt(?, password_hash) AND (login = ? OR email = ?)",
		user.Password, user.Login, user.Email).Select()

	if err != nil {
		// Failing to find an entry is not really an error. It merely means that the
		// authentication failed, so return false in this case.
		if err == pg.ErrNoRows {
			return false, nil
		}
		// Other types of errors have to be logged properly.
		err = errors.Wrapf(err, "database operation error while trying to authenticate user %s", user.Identity())
		return false, err
	}

	// We don't want to return password hash in the password field so we
	// set it to an empty string, which serves two purposes. First, the
	// password hash is not interpreted as password. Second, when using the
	// returned SystemUser instance to update the database the password will
	// remain unmodified in the database. The database treats empty password
	// as an indication that the old password must be preserved.
	user.Password = ""
	return true, err
}

// Fetches a collection of users from the database. The offset and limit specify the
// beginning of the page and the maximum size of the page. If these values are set
// to 0, all users are returned.
func GetUsers(db *dbops.PgDB, offset, limit int, order SystemUserOrderBy) (users SystemUsers, err error) {
	q := db.Model(&users)

	switch order {
	case SystemUserOrderByLoginEmail:
		q = q.OrderExpr("login ASC").OrderExpr("email ASC")
	default:
		q = q.OrderExpr("id ASC")
	}

	q = q.Offset(offset).Limit(limit)
	err = q.Select()

	if err != nil {
		err = errors.Wrapf(err, "problem with fetching a list of users from the database")
	}

	return users, err
}

// Fetches a user with a given id from the database. If the user does not exist
// the nil value is returned.
func GetUserById(db *dbops.PgDB, id int) (*SystemUser, error) {
	user := &SystemUser{Id: id}
	err := db.Select(user)
	if err == pg.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "problem with fetching user %v from the database", id)
	}
	return user, err
}
