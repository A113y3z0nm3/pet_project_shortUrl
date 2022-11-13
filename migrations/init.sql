/*
Таблица с пользователями
*/
CREATE TABLE IF NOT EXISTS cpuser (
    user_id bigserial    NOT NULL PRIMARY KEY,
    username varchar     NOT NULL UNIQUE,
    first_name varchar   NOT NULL DEFAULT '',
    last_name varchar    NOT NULL DEFAULT '',
    password varchar     NOT NULL
);
