# ArestAPI 🎬

A simple REST API for managing movies, built with Go. I made this to learn how to build real backend APIs — user auth, JSON, PostgreSQL, all the good stuff.

It's not fancy, but it works, and it's a good example if you're also learning Go and want to see how a small API is put together.

## What it can do

- Add, view, update, and delete movies
- Register a new user and activate their account with a token
- Log in and get an authentication token
- Rate limiting so people can't spam the API
- Sends activation emails in the background (doesn't block the request)
- Clean JSON responses everywhere

## Tech I used

- **Go** – the whole backend
- **PostgreSQL** – stores all the data
- **httprouter** – handles routing (it's fast and simple)

## Before you start

Make sure you have these installed:

- [Go](https://go.dev/dl/) 1.20 or newer
- [PostgreSQL](https://www.postgresql.org/) running locally (or somewhere you can connect to)

## How to run it

1. Clone the repo:
   ```bash
   git clone https://github.com/zeeshangolang/arestapi.git
   cd arestapi
   ```

2. Set up your database connection (make sure your PostgreSQL is running and you have a database ready)
you can copy the db schema from migration folder that would be easy for you to follow.

3. Run the app:
   ```bash
   go run ./cmd/api
   ```

4. That's it — the API should now be running locally.

## The endpoints

### Movies

| Method | Endpoint          | What it does                | Needs login? |
|--------|-------------------|------------------------------|--------------|
| GET    | `/v1/movies`      | Get all movies               | No           |
| POST   | `/v1/movies`      | Add a new movie              | Yes          |
| GET    | `/v1/movies/:id`  | Get one movie by ID          | Yes          |
| PATCH  | `/v1/movies/:id`  | Update a movie                | Yes          |
| DELETE | `/v1/movies/:id`  | Delete a movie                | Yes          |

### Users & Auth

| Method | Endpoint                     | What it does                          |
|--------|-------------------------------|----------------------------------------|
| POST   | `/v1/users`                  | Register a new user                    |
| PUT    | `/v1/users/activated`        | Activate a user's account with a token |
| POST   | `/v1/tokens/authentication`  | Log in and get an auth token           |

### Health check

| Method | Endpoint            | What it does              |
|--------|----------------------|----------------------------|
| GET    | `/v1/healthcheck`   | Just checks if the API is alive |

## How registering & activating works

1. You send your name, email, and password to `/v1/users`.
2. The API creates your account (not activated yet) and emails you an activation token in the background.
3. You send that token to `/v1/users/activated` to activate your account.
4. Now you can log in at `/v1/tokens/authentication` and start using the protected endpoints (like adding movies).

## How the auth actually works

This isn't JWT — it's simpler than that:

1. When you need a token (for activation or login), the server generates a random string.
2. It sends you that string as your plaintext token.
3. But it only ever saves a **hashed** version of it in the database — never the plaintext.
4. When you send your token back on future requests (`Authorization: Bearer <token>`), the server hashes it and checks for a match in the database.

The nice part: since tokens live in the database, they can be revoked instantly just by deleting the row — something plain JWTs can't do without extra setup.

## Why I built it this way

I wanted to actually understand things like:
- how token-based auth actually works under the hood (this uses random, hashed tokens stored in the database — not JWT) without just using a library that does everything for me
- how background jobs work in Go (like sending emails without slowing down the response)
- how to structure a Go project properly (handlers, models, routes all separated)

It's still a work in progress and I'm adding things as I learn more.

## What's next

- [ ] Add tests
- [ ] Add pagination and filtering for movies
- [ ] Dockerize it
- [ ] Deploy it somewhere live

---

If you're learning Go too and have suggestions, feel free to open an issue or a PR 🙂