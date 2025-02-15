create table if not exists currencies (
    id         smallint     primary key generated always as identity,
    iso_alpha  char(3)      not null unique,
    iso_number smallint     not null unique,
    name       varchar(256) not null unique,
    exponent   smallint     not null,
    constraint currency_exponent_check check ((exponent >= 0)),
    constraint currency_iso_number_check check ((iso_number >= 0))
);

create index if not exists currency_iso_alpha_idx on currencies(iso_alpha);

--------------------------------------------------------------------------------

create table if not exists countries (
    id            smallint     primary key generated always as identity,
    name          text         not null unique,
    iso_number    smallint     not null unique,
    iso2_code     char(2)      not null unique,
    iso3_code     char(3)      not null unique,
    fips_code     char(2)      not null,
    geonameid     integer,
    e164_code     smallint     not null,
    tld           char(2)      not null,
    dialing_code  varchar(32)  not null,
    continent     varchar(16)  not null,
    capital       varchar(256) not null,
    capital_tz    varchar(64)  not null,
    area_km2      integer      not null,
    currency_id   smallint     not null references currencies(id) on delete restrict,
    constraint country_area_check check ((area_km2 >= 0)),
    constraint country_e164_code_check check ((e164_code >= 0)),
    constraint country_geonameid_check check ((geonameid >= 0)),
    constraint country_iso_number_check check ((iso_number >= 0))
);

create index if not exists country_iso_number_idx on countries (iso_number);
create index if not exists country_iso2_code_idx  on countries (iso2_code);
create index if not exists country_iso3_code_idx  on countries (iso3_code);

--------------------------------------------------------------------------------

create table if not exists water_types (
    id          integer      primary key generated always as identity,
    name        varchar(32)  not null unique,
    description varchar(256) not null default '',
    density     numeric(4,3) not null
);

--------------------------------------------------------------------------------

create table if not exists water_bodies (
    id          integer      primary key generated always as identity,
    name        varchar(32)  not null unique,
    description varchar(256) not null default ''
);

--------------------------------------------------------------------------------

create table if not exists dive_sites (
    id            bigint        primary key generated always as identity,
    version       integer       not null default 1,
    created_at    timestamp(8) with time zone not null default now(),
    updated_at    timestamp(8) with time zone not null default now(),
    owner_id      varchar(32)   not null,
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

