package main

import (
	"errors"
	"fmt"
	"github.com/ol-ilyassov/spa_final/internal/data"
	"github.com/ol-ilyassov/spa_final/internal/validator"
	"net/http"
)

func (app *application) createMusicHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string `json:"title"`
		Year   int32  `json:"year"`
		Author string `json:"author"`
		Link   string `json:"link"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	music := &data.Music{
		Title:  input.Title,
		Year:   input.Year,
		Author: input.Author,
		Link:   input.Link,
	}

	v := validator.New()
	if data.ValidateMusic(v, music); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Musics.Insert(music)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/musics/%d", music.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"music": music}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMusicHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	music, err := app.models.Musics.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"music": music}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMusicHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	music, err := app.models.Musics.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title  *string `json:"title"`
		Year   *int32  `json:"year"`
		Author *string `json:"author"`
		Link   *string `json:"link"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Year != nil {
		music.Year = *input.Year
	}
	if input.Author != nil {
		music.Author = *input.Author
	}
	if input.Link != nil {
		music.Link = *input.Link
	}

	v := validator.New()

	if data.ValidateMusic(v, music); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Musics.Update(music)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"music": music}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteMusicHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Musics.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "music successgully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMusicsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Author string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Author = app.readString(qs, "author", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "-id", "-title", "-year"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	musics, metadata, err := app.models.Musics.GetAll(input.Title, input.Author, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"musics": musics, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
