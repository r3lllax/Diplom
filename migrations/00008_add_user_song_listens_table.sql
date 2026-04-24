-- +goose Up
-- +goose StatementBegin
create table user_songs_listens(
    id serial primary key,
    user_id int not null,
    song_id int not null,
    listens int not null,
    last_listen_time timestamp,
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (song_id) references songs(id) on delete cascade

)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table user_songs_listens
-- +goose StatementEnd
