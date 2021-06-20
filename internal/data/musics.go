package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ol-ilyassov/spa_final/internal/validator"
	"time"
)

type Music struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Author    string    `json:"genres"`
	Link      string    `json:"link,omitempty"`
	Version   int32     `json:"version"`
}

type MusicModel struct {
	DB *sql.DB
}

func ValidateMusic(v *validator.Validator, music *Music) {
	v.Check(music.Title != "", "title", "must be provided")
	v.Check(len(music.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(music.Year != 0, "year", "must be provided")
	v.Check(music.Year >= 1888, "year", "must be greater than 1888")
	v.Check(music.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(music.Author != "", "author", "must be provided")
	v.Check(len(music.Author) <= 300, "author", "must not be more than 300 bytes long")
	v.Check(music.Link != "", "link", "must be provided")
	v.Check(len(music.Link) <= 500, "link", "must not be more than 500 bytes long")

}

func (m MusicModel) Insert(music *Music) error {
	query := `
INSERT INTO musics (title, year, author, link) 
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version`

	args := []interface{}{music.Title, music.Year, music.Author, music.Link}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&music.ID, &music.CreatedAt, &music.Version)
}

func (m *MusicModel) Get(id int64) (*Music, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
SELECT id, created_at, title, year, author, version
FROM musics
WHERE id = $1`

	var music Music
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&music.ID,
		&music.CreatedAt,
		&music.Title,
		&music.Year,
		&music.Author,
		&music.Link,
		&music.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &music, nil
}

func (m *MusicModel) Update(music *Music) error {
	query := `
UPDATE musics
SET title = $1, year = $2, author = $3, link = $4,version = version + 1
WHERE id = $4 AND version = $5
RETURNING version`
	args := []interface{}{
		music.Title,
		music.Year,
		music.Author,
		music.Link,
		music.ID,
		music.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&music.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m *MusicModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM musics WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m *MusicModel) GetAll(title, author string, filters Filters) ([]*Music, Metadata, error) {
	query := fmt.Sprintf(`
SELECT count(*) OVER(), id, created_at, title, year, author, link, version
FROM musics
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (to_tsvector('simple', author) @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY %s %s, id ASC
LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, author, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	musics := []*Music{}

	for rows.Next() {
		var music Music
		err := rows.Scan(
			&totalRecords,
			&music.ID,
			&music.CreatedAt,
			&music.Title,
			&music.Year,
			&music.Author,
			&music.Link,
			&music.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		musics = append(musics, &music)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return musics, metadata, nil
}
