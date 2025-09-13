create table if not exists users (
    id              bigint       primary key generated always as identity,
    created_at      timestamp(6) with time zone not null default now(),
    updated_at      timestamp(6) with time zone not null default now(),
    name            varchar(256) not null,
    email           varchar(256) not null unique,
    hashed_password char(60)     not null
);

create index if not exists user_email_idx on users (email);

insert into users (created_at, updated_at, name, email, hashed_password) values (
    '2001-02-03 04:05:06',
    '2011-12-13 14:15:16',
    'Alice Jones',
    'alice@example.com',
    '$2a$12$NuTjWXm3KKntReFwyBVHyuf/to.HEwTy.eS206TNfkGfr6HzGJSWG'
);

