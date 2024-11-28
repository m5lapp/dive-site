create table if not exists users (
    id              bigint       primary key generated always as identity,
    created_at      timestamp(8) with time zone not null default now(),
    updated_at      timestamp(8) with time zone not null default now(),
    name            varchar(256) not null,
    email           varchar(256) not null unique,
    hashed_password char(60)     not null
);

create index if not exists user_email_idx on users (email);

