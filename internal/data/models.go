package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Musics      MusicModel
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Musics:      MusicModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}
