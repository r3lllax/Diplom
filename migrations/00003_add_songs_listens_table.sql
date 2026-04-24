-- +goose Up
-- +goose StatementBegin
create table songs_listens(
    id serial primary key unique,
    song_id int not null,
    listens int not null,
    foreign key (song_id) references songs(id) on delete cascade
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table songs_listens
-- +goose StatementEnd
