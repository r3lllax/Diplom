-- +goose Up
-- +goose StatementBegin
create table liked_playlists(
    id serial primary key,
    user_id int not null,
    playlist_id int not null,
    liked_at timestamp,
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (playlist_id) references playlists(id) on delete cascade
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table liked_playlists
-- +goose StatementEnd
