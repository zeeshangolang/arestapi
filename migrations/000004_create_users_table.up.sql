CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    createdat timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated boolean NOT NULL,
    version integer NOT NULL DEFAULT 1 
);