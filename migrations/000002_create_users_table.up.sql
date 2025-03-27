create table if not exists users (
    id              bigint       primary key generated always as identity,
    created_at      timestamp(8) with time zone not null default now(),
    updated_at      timestamp(8) with time zone not null default now(),
    name            varchar(256) not null,
    friendly_name   varchar(256) not null,
    email           varchar(256) not null unique,
    hashed_password char(60)     not null,
    suspended       bool         not null default false,
    deleted         bool         not null default false,
    dark_mode       boolean      not null default true,
    diving_since    date         not null default '0001-01-01'::date,
    dive_number_offset smallint  not null default 0,
    default_diving_country_id integer not null references countries(id) on delete restrict,
    default_diving_tz varchar(64) not null default 'Etc/UTC'
);

create index if not exists user_email_idx on users (email);

--------------------------------------------------------------------------------

create table if not exists sessions (
    token  text      primary key,
    data   bytea     not null,
    expiry timestamp with time zone not null
);

create index sessions_expiry_idx on sessions(expiry);

