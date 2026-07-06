package main

import (
	//"encoding/json"

	"errors"
	"fmt"
	"net/http"

	"main/internal/data"
	"main/internal/validator"
)

func (app *application) createmovie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int          `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genre   []string     `json:"genres"`
	}

	err := app.Readjson(w, r, &input)
	if err != nil {
		app.badRequestresponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    int32(input.Year),
		Runtime: input.Runtime,
		Genre:   input.Genre,
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("v1/movies/%d", movie.Id))

	err = app.writejson(w, http.StatusCreated, envolpe{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	v := validator.New()

	if data.ValidateMovie(v, *movie); !v.Valid() {
		app.Failedvalidationerror(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) viewmovie(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ERRRecodrdNotFound):
			app.notFound(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writejson(w, http.StatusOK, envolpe{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	//fmt.Fprintf(w, "RESULT FOR %d", id)
}

func (app *application) Updatemovie(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ERRRecodrdNotFound):
			app.notFound(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  *[]string     `json:"genres"`
	}

	err = app.Readjson(w, r, &input)
	if err != nil {
		app.badRequestresponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genre = *input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, *movie); !v.Valid() {
		app.Failedvalidationerror(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ERREditConfilct):
			app.editConflictError(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writejson(w, 200, envolpe{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) DeletemovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w, r)
		return
	}
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ERRRecodrdNotFound):
			app.notFound(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return

	}

	err = app.writejson(w, 200, envolpe{"movie": "movie deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listMovieshandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.csvReader(qs, "genres", []string{})
	input.Filters.Page = app.readint(qs, "page", 1, v)
	input.Filters.Page_size = app.readint(qs, "page_size", 19, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.Failedvalidationerror(w, r, v.Errors)
		return
	}

	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	err = app.writejson(w, 200, envolpe{"movie": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
