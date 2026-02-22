create table if not exists dive_plans (
    id           bigint        primary key generated always as identity,
    version      integer       not null default 1,
    created_at   timestamp(6) with time zone not null default now(),
    updated_at   timestamp(6) with time zone not null default now(),
    owner_id     bigint        not null references users(id) on delete restrict,
    name         varchar(256)  not null,
    notes        text          not null default '',
    is_solo_dive boolean       not null,
    descent_rate numeric(3, 1) not null,
    ascent_rate  numeric(3, 1) not null,
    sac_rate     numeric(4, 1) not null,
    tank_count   smallint      not null,
    tank_volume  numeric(3, 1) not null,
    working_pressure smallint  not null,
    dive_factor  numeric(2, 1) not null,
    fn2          numeric(4, 3) not null default 0.0,
    fhe          numeric(4, 3) not null default 0.0,
    max_ppo2     numeric(3, 2) not null
);

create trigger update_updated_at_timestamp
before update on dive_plans
for each row execute function update_updated_at_timestamp();

create index if not exists dive_plans_owner_id_idx on dive_plans (owner_id);

--------------------------------------------------------------------------------

create table if not exists dive_plan_stops (
    id            bigint        primary key generated always as identity,
    dive_plan_id  bigint        not null references dive_plans(id) on delete cascade,
    sort          smallint      not null,
    depth         numeric(4, 1) not null,
    duration      numeric(4, 1) not null,
    is_transition boolean       not null default false,
    comment       varchar(256)  not null default ''
);

create index if not exists dive_plans_owner_id_idx on dive_plans (owner_id);

