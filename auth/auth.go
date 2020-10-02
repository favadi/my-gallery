package auth

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("username/password combination doesn't match")

type Authenticator struct {
	db *sqlx.DB
}

func NewAuthenticator(db *sqlx.DB) *Authenticator {
	return &Authenticator{db: db}
}

type UserDB struct {
	User
	PasswordHard string `db:"password_hash"`
}

type User struct {
	ID       int64
	Username string
	FullName string `db:"full_name"`
	Created  time.Time
	Updated  time.Time
}

func (a *Authenticator) Authenticate(username, password string) (User, error) {
	const query = `SELECT username, password_hash, full_name, created, updated
FROM users
WHERE username = $1;`
	var user UserDB
	err := a.db.Get(&user, query, username)
	if err == sql.ErrNoRows {
		return User{}, ErrInvalidCredentials
	} else if err != nil {
		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHard), []byte(password)); err != nil {
		log.Printf("failed to authenticate, password doesn't match: username=%s", username)
		return User{}, ErrInvalidCredentials
	}
	return user.User, nil
}
