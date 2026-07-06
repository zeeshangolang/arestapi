package main

import (
	"context"
	"main/internal/data"
	"net/http"
)

type ContextKey string

const UserContext = ContextKey("user")

func (app *application) SetUserContext(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContext, user)
	return r.WithContext(ctx)

}

func (app *application) GetUserContext(r *http.Request) *data.User {
	user, ok := r.Context().Value(UserContext).(*data.User)
	if !ok {
		panic("missing  user details")
	}
	return user
}
