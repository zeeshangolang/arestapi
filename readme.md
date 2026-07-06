# ArestAPI

A lightweight, robust RESTful API for managing a movie database, built with Go.

## Features

- **RESTful Endpoints**: Perform CRUD operations on movie records.
- **Efficient Data Handling**: Uses Go's standard library and PostgreSQL for reliable storage.
- **JSON-first Design**: Clean, standard JSON responses for easy integration.

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) (1.20 or higher)
- [PostgreSQL](https://www.postgresql.org/)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/zeeshangolang/arestapi.git
   cd arestapi

   ##Run the application:
    go run ./cmd/api

   API Reference
Endpoint    	Method            	Description
/movies	      GET              	  Retrieve all movies
/movies	      POST	              Add a new movie
/movies/:id  	GET               	Retrieve a single movie by ID
/movies/:id	  DELETE	            Remove a movie reco