create or replace function update_updated_at_timestamp()
returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language 'plpgsql';

--------------------------------------------------------------------------------

create table if not exists users (
    id              bigint       primary key generated always as identity,
    version         integer      not null default 1,
    created_at      timestamp(6) with time zone not null default now(),
    updated_at      timestamp(6) with time zone not null default now(),
    name            varchar(256) not null,
    friendly_name   varchar(256) not null,
    email           varchar(256) not null unique,
    hashed_password char(60)     not null,
    suspended       bool         not null default false,
    deleted         bool         not null default false,
    last_log_in     timestamp(6) with time zone not null default '0001-01-01T00:00:00.000'::date,
    dark_mode       boolean      not null default true,
    diving_since    date         not null default '0001-01-01'::date,
    dive_number_offset smallint  not null default 0,
    default_diving_country_id integer not null references countries(id) on delete restrict,
    default_diving_tz varchar(64) not null default 'Etc/UTC'
);

create trigger update_updated_at_timestamp
before update on users
for each row execute function update_updated_at_timestamp();

create index if not exists user_email_idx on users (email);

--------------------------------------------------------------------------------

create table if not exists sessions (
    token  text      primary key,
    data   bytea     not null,
    expiry timestamp with time zone not null
);

create index sessions_expiry_idx on sessions(expiry);

