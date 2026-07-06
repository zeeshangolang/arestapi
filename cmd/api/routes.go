package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	//"golang.org/x/net/route"
)

func (app *application) Routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFound)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", (app.listMovieshandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.RequireAuthentication(app.createmovie))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.RequireAuthentication(app.viewmovie))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.RequireAuthentication(app.Updatemovie))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.RequireAuthentication(app.DeletemovieHandler))
	router.HandlerFunc(http.MethodPost, "/v1/users", app.RegisterUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.ActivateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.AuntenticateTokenHandler)
	return app.RecoverPanic(app.NewrateLimiter(app.Authenticate(router)))
}
