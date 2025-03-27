create table if not exists water_types (
    id          integer      primary key generated always as identity,
    name        varchar(32)  not null unique,
    description varchar(256) not null default '',
    density     numeric(4,3) not null
);

--------------------------------------------------------------------------------

insert into water_types (name, description, density) values
    ('Fresh Water', 'Water with low concentrations of dissolved salts and other dissolved solids', 1.000),
    ('Salt Water', 'Water from a sea or ocean', 1.025);

--------------------------------------------------------------------------------

create table if not exists water_bodies (
    id          integer      primary key generated always as identity,
    name        varchar(32)  not null unique,
    description varchar(256) not null default ''
);

--------------------------------------------------------------------------------

insert into water_bodies (name, description) values
    ('Artificial Lake', 'A man-made lake'),
    ('Harbour', 'An artificial or naturally occurring body of water where ships are stored or may shelter from the ocean''s weather and currents'),
    ('Lake', 'A relatively large body of water contained on a body of land'),
    ('Ocean', 'A major body of salty water'),
    ('River', 'A natural waterway that flows from higher ground to lower ground'),
    ('Quarry', 'A dive in an old quarry');

--------------------------------------------------------------------------------

create table if not exists dive_sites (
    id            bigint        primary key generated always as identity,
    version       integer       not null default 1,
    created_at    timestamp(8) with time zone not null default now(),
    updated_at    timestamp(8) with time zone not null default now(),
    owner_id      bigint        not null references users(id) on delete restrict,
    name          varchar(256)  not null,
    alt_name      varchar(256)  not null default '',
    location      varchar(256)  not null,
    region        varchar(256)  not null default '',
    country_id    integer       not null references countries(id) on delete restrict,
    timezone      varchar(64)   not null,
    latitude      numeric(8,6),
    longitude     numeric(9,6),
    water_body_id integer       not null references water_bodies(id) on delete restrict,
    water_type_id integer       not null references water_types(id)  on delete restrict,
    altitude      smallint      not null default 0,
    max_depth     numeric(4,1),
    notes         text          not null default '',
    rating        smallint,
    constraint dive_site_rating_check check ((rating >= 0))
);

create index if not exists dive_site_name_idx
    on countries using gin (to_tsvector('simple', name));

create index if not exists dive_site_country_idx on dive_sites (country_id);

