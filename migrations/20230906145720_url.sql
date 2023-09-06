-- +goose Up

-- +goose StatementBegin

CREATE TABLE
    url (
        id VARCHAR(255) NOT NULL PRIMARY KEY,
        shorturl VARCHAR(255) NOT NULL UNIQUE,
        originalurl VARCHAR(255) NOT NULL
    );

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE IF EXISTS url;

-- +goose StatementEnd