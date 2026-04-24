-- +goose Up
-- +goose StatementBegin
create table playlists(
    id serial primary key,
    user_id int not null,
    title varchar(255) not null,
    description varchar(255),
    volume_path text not null,
    foreign key (user_id) references users(id) on delete cascade,
    is_private boolean default false,
    is_available boolean default false
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table playlists
-- +goose StatementEnd
