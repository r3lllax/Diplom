-- +goose Up
-- +goose StatementBegin
create table songs(
    id serial primary key unique,
    user_id int not null,
    author varchar(255) not null,
    name varchar(255) not null,
    duration int not null,
    file_path text not null unique,
    volume_path text not null unique,
    uploaded_at timestamp,
    is_available boolean default true,
    /*TODO: is_private boolean default false, */
    foreign key (user_id) references users(id) on delete cascade
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table songs
-- +goose StatementEnd
