package main

import (
	"errors"
	"main/internal/data"
	"main/internal/validator"
	"net/http"
	"time"
)

func (app *application) AuntenticateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.Readjson(w, r, &input)
	if err != nil {
		app.badRequestresponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePlainPassword(v, input.Password)

	if !v.Valid() {
		app.Failedvalidationerror(w, r, v.Errors)
		return
	}

	user, err := app.models.User.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ERRRecodrdNotFound):
			app.InvalidCredentialsRes(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.InvalidCredentialsRes(w, r)
		return
	}

	token, err := app.models.Token.New(user.ID, 1*24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writejson(w, http.StatusCreated, envolpe{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
