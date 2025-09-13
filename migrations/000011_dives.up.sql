create table if not exists dives (
    id               bigint        primary key generated always as identity,
    version          integer       not null default 1,
    created_at       timestamp(6)  with time zone not null default now(),
    updated_at       timestamp(6)  with time zone not null default now(),
    owner_id         bigint        not null references users(id) on delete cascade,
    number           integer       not null,
    activity         varchar(256)  not null,
    dive_site_id     bigint        not null references dive_sites(id) on delete restrict,
    operator_id      bigint            null references operators(id) on delete restrict,
    price            numeric(13, 3)    null,
    currency_id      smallint          null references currencies(id) on delete restrict,
    trip_id          bigint            null references trips(id) on delete set null,
    certification_id bigint            null references certifications(id) on delete set null,
    date_time_in     timestamp(6) with time zone not null,
    max_depth        numeric(4, 1) not null,
    avg_depth        numeric(4, 1)     null,
    bottom_time      bigint        not null,
    safety_stop      bigint            null,
    water_temp       smallint          null,
    air_temp         smallint          null,
    visibility       numeric(3, 1)     null,
    current_id       smallint          null references currents(id) on delete restrict,
    waves_id         smallint          null references waves(id) on delete restrict,
    buddy_id         bigint            null references buddies(id) on delete restrict,
    buddy_role_id    smallint          null references buddy_roles(id) on delete restrict,
    weight_used      numeric(4, 2)     null,
    weight_notes     varchar(1024) not null default '',
    equipment_notes  varchar(1024) not null default '',
    tank_configuration_id smallint not null references tank_configurations(id) on delete restrict,
    tank_material_id smallint      not null references tank_materials(id) on delete restrict,
    tank_volume      numeric(4, 2) not null,
    gas_mix_id       smallint      not null references gas_mixes(id) on delete restrict,
    fo2              numeric(4, 3) not null default 0.21,
    pressure_in      smallint          null,
    pressure_out     smallint          null,
    gas_mix_notes    varchar(1024) not null default '',
    entry_point_id   smallint      not null references entry_points(id) on delete restrict,
    rating           smallint          null,
    notes            text          not null default '',
    unique (owner_id, number)
);

create trigger update_updated_at_timestamp
before update on dives
for each row execute function update_updated_at_timestamp();

create index if not exists dives_owner_id_idx on dives (owner_id);

--------------------------------------------------------------------------------

create table if not exists dive_equipment (
    dive_id      bigint   not null references dives(id) on delete cascade,
    equipment_id smallint not null references equipment(id) on delete restrict,
    primary key(dive_id, equipment_id)
);

create index if not exists dive_equipment_dive_id_idx
    on dive_equipment (dive_id);

--------------------------------------------------------------------------------

create table if not exists dive_dive_properties (
    dive_id     bigint   not null references dives(id) on delete cascade,
    property_id smallint not null references dive_properties(id) on delete restrict,
    primary key(dive_id, property_id)
);

create index if not exists dive_dive_properties_dive_id_idx
    on dive_dive_properties (dive_id);

