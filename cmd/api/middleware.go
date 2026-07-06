package main

import (
	"errors"
	"fmt"
	"main/internal/data"
	"main/internal/validator"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

//	func (app *application) globallimiter(next http.Handler) http.Handler {
//		limiter := rate.NewLimiter(1, 20)
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			if !limiter.Allow() {
//				app.rattLimitExceded(w, r)
//				return
//			}
//			next.ServeHTTP(w, r)
//		})
//	}
func (app *application) NewrateLimiter(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.config.limiter.enabeld {

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{

					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rattLimitExceded(w, r)
				return
			}

			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Vary", "Authorization")

		AuthorizationHeader := r.Header.Get("Authorization")

		if AuthorizationHeader == "" {
			r = app.SetUserContext(r, data.AnonymusUser)
			next.ServeHTTP(w, r)
			return
		}

		headerparts := strings.Split(AuthorizationHeader, " ")
		if len(headerparts) != 2 || headerparts[0] != "Bearer" {
			app.InvalidauthenticationTokenResponse(w, r)
			return
		}

		token := headerparts[1]

		v := validator.New()
		if data.ValidatePlainPassword(v, token); !v.Valid() {
			app.InvalidauthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.User.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ERRRecodrdNotFound):
				app.InvalidauthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.SetUserContext(r, user)

		next.ServeHTTP(w, r)

	})
}

func (app *application) RequireAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.GetUserContext(r)

		if user.IsAnonymus() {
			app.AuthenticationRequiredresponse(w, r)
			fmt.Println("this the first one ")
			return
		}

		if !user.Activated {
			app.InActiveAccountResponse(w, r)
			fmt.Println("this one ran")
			return
		}

		next.ServeHTTP(w, r)
	})
}
