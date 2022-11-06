/*
Таблица с пользователями
*/
CREATE TABLE IF NOT EXISTS cpuser (
    user_id bigserial    NOT NULL PRIMARY KEY,
    username varchar     NOT NULL UNIQUE,
    first_name varchar   NOT NULL DEFAULT '',
    last_name varchar    NOT NULL DEFAULT '',
    subscribe role       NOT NULL,
    password varchar     NOT NULL
);

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'stage') THEN
            CREATE TYPE stage AS ENUM
                (
                    'daily', 'monthly', 'morning_shift', 'evening_shift'
                    );
        END IF;
    END$$;
