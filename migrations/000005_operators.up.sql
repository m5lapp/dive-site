create table if not exists operator_types (
    id          smallint     primary key generated always as identity,
    name        varchar(64)  not null unique,
    description varchar(128) not null
);

--------------------------------------------------------------------------------

insert into operator_types (name, description) values
    ('Dive Club', 'An organised dive club'),
    ('Dive School', 'A dive school that teaches courses'),
    ('Dive Shop', 'A dive equipment shop that also runs diving activities'),
    ('Other', 'Miscellaneous');

--------------------------------------------------------------------------------

create table if not exists operators (
    id               bigint        primary key generated always as identity,
    created_at       timestamp(6) with time zone not null default now(),
    updated_at       timestamp(6) with time zone not null default now(),
    owner_id         bigint        not null references users(id)          on delete restrict,
    operator_type_id smallint      not null references operator_types(id) on delete restrict,
    name             varchar(256)  not null,
    street           varchar(256)  not null default '',
    suburb           varchar(256)  not null default '',
    state            varchar(256)  not null default '',
    postcode         varchar(16)   not null default '',
    country_id       smallint      not null references countries(id)      on delete restrict,
    website_url      varchar(2048) not null default '',
    email_address    varchar(254)  not null default '',
    phone_number     varchar(32)   not null default '',
    comments         text          not null default ''
);

create index if not exists operators_name_idx
    on operators using gin (to_tsvector('simple', name));

create index if not exists operators_owner_id_idx on operators (owner_id);

