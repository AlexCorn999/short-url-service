-- +goose Up

-- +goose StatementBegin

CREATE TABLE users ( id serial primary key);

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

DROP TABLE users;

-- +goose StatementEnd