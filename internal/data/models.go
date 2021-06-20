package data

import (
	"database/sql"
	"errors"
)

// Return this error at Get() when movie doesn't exist in db.
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Stores and wraps all models.
type Models struct {
	Movies MovieModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
