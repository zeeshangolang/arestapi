package data

import (
	"database/sql"
	"errors"
)

var (
	ERRRecodrdNotFound = errors.New("RECORD NOT FOUND")
	ERREditConfilct    = errors.New("edit conflict")
)

type Models struct {
	Movies MovieModel
	User   UserModel
	Token  TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		User:   UserModel{DB: db},
		Token:  TokenModel{DB: db},
	}
}
