-- migrations/000001_create_users_table.up.sql
CREATE TYPE user_type AS ENUM
('regular', 'guest');

CREATE TABLE users
(
    id UUID PRIMARY KEY,
    email TEXT UNIQUE,
    password_hash TEXT,
    name TEXT,
    user_type user_type NOT NULL,
    created_at TIMESTAMP
    WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP
    WITH TIME ZONE NOT NULL
);

    CREATE INDEX idx_users_email ON users(email);

    -- migrations/000001_create_users_table.down.sql
    DROP TABLE IF EXISTS users;
    DROP TYPE IF EXISTS user_type;