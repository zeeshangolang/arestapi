package main

import (
	//"crypto/internal/edwards25519/field"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"main/internal/validator"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

type envolpe map[string]any

func (app *application) writejson(w http.ResponseWriter, status int, data envolpe, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, values := range headers {
		w.Header()[key] = values
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(js))

	return nil
}

func (app *application) Readjson(w http.ResponseWriter, r *http.Request, dst any) error {

	maxBytes := 1_048_576

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {
		var SyntaxError *json.SyntaxError
		var unmarshaltypeerror *json.UnmarshalTypeError
		var Invalidunmarshalerror *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &SyntaxError):
			return fmt.Errorf("Body contains Invalid Error at char %d", SyntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("Error unexpexted end of file ")

		case errors.As(err, &unmarshaltypeerror):
			if unmarshaltypeerror.Field != "" {
				return fmt.Errorf("Field is wrong at %q", unmarshaltypeerror.Field)
			}
			return fmt.Errorf("Field is  really wrong at %d", unmarshaltypeerror.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("request must not be emoty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldname := strings.TrimPrefix(err.Error(), "json: unknown field ")

			return fmt.Errorf("body contains unknown key %s", fieldname)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &Invalidunmarshalerror):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("must not be empty ")

	}
	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) csvReader(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) readint(qs url.Values, key string, defaultValue int, v *validator.Validator) int {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be a valid value")
		return defaultValue
	}
	return i

}

func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {

		defer func() {
			defer app.wg.Done()
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		fn()
	}()
}

// Title
// Genres
// Page
// string
// []string
// int
// PageSize int
// Sort string
