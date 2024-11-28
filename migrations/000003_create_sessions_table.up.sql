create table if not exists sessions (
    token  text      primary key,
    data   bytea     not null,
    expiry timestamp with time zone not null
);

create index sessions_expiry_idx on sessions(expiry);

