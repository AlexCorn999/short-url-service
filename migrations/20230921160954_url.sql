-- +goose Up

-- +goose StatementBegin

CREATE TABLE
    url (
        id SERIAL PRIMARY KEY,
        shorturl VARCHAR(255) NOT NULL UNIQUE,
        originalurl VARCHAR(255) NOT NULL,
        user_id integer REFERENCES users (id)
    );

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS url;

-- +goose StatementEnd