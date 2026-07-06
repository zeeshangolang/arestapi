package main

import (
	"errors"
	"fmt"
	"time"

	"main/internal/data"
	_ "main/internal/mailer"
	"main/internal/validator"
	"net/http"
)

func (app *application) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.Readjson(w, r, &input)
	if err != nil {
		app.badRequestresponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.Failedvalidationerror(w, r, v.Errors)
		return
	}

	err = app.models.User.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrrDuplicateEmail):
			v.AddError("email", "user with this email already exist")
			app.Failedvalidationerror(w, r, v.Errors)

		default:
			app.serverErrorResponse(w, r, err)
		}

		return

	}

	token, err := app.models.Token.New(user.ID, 1*60*time.Hour, data.ScopeAvctivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {

		data := map[string]any{
			"activationtoken": token.Plaintext,
			"ID":              user.ID,
		}

		err := app.mailer.Send(user.Email, "user_template.html", data)
		if err != nil {
			app.logger.Error(fmt.Sprintf("%v", err))
		}
	})

	err = app.writejson(w, http.StatusAccepted, envolpe{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.Readjson(w, r, &input)
	if err != nil {
		app.badRequestresponse(w, r, err)
		return
	}
	v := validator.New()
	if data.CheckPlainText(v, input.TokenPlaintext); !v.Valid() {
		app.Failedvalidationerror(w, r, v.Errors)
		return
	}

	user, err := app.models.User.GetForToken(data.ScopeAvctivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ERRRecodrdNotFound):
			v.AddError("token", "invalid or expired token")
			app.Failedvalidationerror(w, r, v.Errors)

		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true
	err = app.models.User.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ERREditConfilct):
			app.editConflictError(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Token.DeleteForAllUsers(data.ScopeAvctivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writejson(w, 200, envolpe{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
