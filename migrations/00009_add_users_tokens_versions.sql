-- +goose Up
-- +goose StatementBegin
create table users_tokens_versions(
    id serial primary key unique,
    user_id int not null,
    token_version int not null,
    foreign key (user_id) references users(id) on delete cascade
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users_tokens_versions
-- +goose StatementEnd
