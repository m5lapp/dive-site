create table if not exists water_types (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into water_types (sort, is_default, name, description) values
    (10, false, 'Fresh Water', 'Water with low concentrations of dissolved salts and other dissolved solids'),
    (20, true,  'Salt Water', 'Water from a sea or ocean');

--------------------------------------------------------------------------------

create table if not exists water_bodies (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into water_bodies (sort, is_default, name, description) values
    (10, false, 'Artificial Lake', 'A man-made lake'),
    (20, false, 'Harbour', 'An artificial or naturally occurring body of water where ships are stored or may shelter from the ocean''s weather and currents'),
    (30, false, 'Lake', 'A relatively large body of water contained on a body of land'),
    (40, false, 'Ocean', 'An immense, continuous expanse of salt water with continents acting as islands within it'),
    (50, false, 'River', 'A natural waterway that flows from higher ground to lower ground'),
    (60, false, 'Sea', 'A large body of saltwater that partially borders a landmass'),
    (70, false, 'Quarry', 'A dive in an old quarry');

--------------------------------------------------------------------------------

create table if not exists dive_sites (
    id            bigint        primary key generated always as identity,
    version       integer       not null default 1,
    created_at    timestamp(6) with time zone not null default now(),
    updated_at    timestamp(6) with time zone not null default now(),
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

create trigger update_updated_at_timestamp
before update on dive_sites
for each row execute function update_updated_at_timestamp();

create index if not exists dive_site_name_idx
    on countries using gin (to_tsvector('simple', name));

create index if not exists dive_site_country_idx on dive_sites (country_id);

