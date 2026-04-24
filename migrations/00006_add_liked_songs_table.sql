-- +goose Up
-- +goose StatementBegin
create table liked_songs(
    id serial primary key,
    user_id int not null,
    song_id int not null,
    liked_at timestamp,
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (song_id) references songs(id) on delete cascade
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table liked_songs
-- +goose StatementEnd
