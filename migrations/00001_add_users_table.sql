-- +goose Up
-- +goose StatementBegin
create table users(
    id serial primary key unique,
    name varchar(30) not null,
    email varchar(255) unique not null,
    password varchar(72) not null,
    is_private boolean default false,
    photo_file text not null,
    registrated_at timestamp
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users
-- +goose StatementEnd
