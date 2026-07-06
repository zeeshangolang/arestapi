package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"main/internal/validator"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrrDuplicateEmail = errors.New("error duplicate email")
	ErrRecNotFound     = errors.New("record not found ")
)

type UserModel struct {
	DB *sql.DB
}

var AnonymusUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	Createdat time.Time `json:"createdat"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymus() bool {
	return u == AnonymusUser
}

type password struct {
	plaintext *string
	hash      []byte
}

func (m UserModel) Insert(user *User) error {
	stmt := `INSERT INTO users 	(name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, createdat, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.Createdat, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(Email string) (*User, error) {
	stmt := `SELECT id, createdat, name, email, password_hash, activated, version 
	FROM users 
	WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, Email).Scan(
		&user.ID,
		&user.Createdat,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ERRRecodrdNotFound
		default:
			return nil, err
		}
	}
	return &user, nil

}

func (m UserModel) Update(user *User) error {
	query := `
	UPDATE 	users 
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "user_email_key"`:
			return ErrrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ERREditConfilct
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlainText string) (*User, error) {
	Tokenhash := sha256.Sum256([]byte(tokenPlainText))

	stmt := `SELECT users.id , users.createdat , users.name, users.email, users.password_hash, users.activated, users.version 
	FROM users INNER JOIN tokens ON users.id = tokens.user_id 
	WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expiry > $3`

	args := []any{Tokenhash[:], tokenScope, time.Now()}

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(
		&user.ID,
		&user.Createdat,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ERRRecodrdNotFound
		default:
			return nil, err
		}

	}

	return &user, nil
}

func (p *password) Set(PlaintextPass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(PlaintextPass), 13)
	if err != nil {
		return err

	}

	p.plaintext = &PlaintextPass
	p.hash = hash

	return nil

}

func (p *password) Matches(PlaintextPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(PlaintextPass))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "email must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a vlid email")

}

func ValidatePlainPassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be 8 bytes long ")
	v.Check(len(password) <= 72, "password", "lss than 72 butea")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "name must be provided ")
	v.Check(len(user.Name) <= 500, "name", "less than 500 ")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePlainPassword(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("no hash")
	}
}
