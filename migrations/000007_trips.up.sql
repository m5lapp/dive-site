create table if not exists trips (
    id          bigint        primary key generated always as identity,
    created_at  timestamp(6)  with time zone not null default now(),
    updated_at  timestamp(6)  with time zone not null default now(),
    owner_id    bigint        not null references users(id) on delete cascade,
    name        varchar(256)  not null,
    start_date  timestamp(6)  with time zone not null default now(),
    end_date    timestamp(6)  with time zone not null default now(),
    description varchar(1024) not null default '',
    rating      smallint,
    operator_id bigint                 references operators(id) on delete restrict,
    price       numeric(13, 3),
    currency_id smallint               references currencies(id) on delete restrict,
    notes       text          not null default ''
);

create trigger update_updated_at_timestamp
before update on trips
for each row execute function update_updated_at_timestamp();

create index if not exists trips_owner_id_idx on trips (owner_id);

--------------------------------------------------------------------------------

create table if not exists certifications (
    id            bigint        primary key generated always as identity,
    created_at    timestamp(6)  with time zone not null default now(),
    updated_at    timestamp(6)  with time zone not null default now(),
    owner_id      bigint        not null references users(id) on delete cascade,
    course_id     smallint      not null references agency_courses(id) on delete restrict,
    start_date    timestamp(6)  with time zone not null default now(),
    end_date      timestamp(6)  with time zone not null default now(),
    operator_id   bigint        not null references operators(id) on delete restrict,
    instructor_id bigint        not null references buddies(id) on delete restrict,
    price         numeric(13, 3),
    currency_id   smallint               references currencies(id) on delete restrict,
    rating        smallint,
    notes         text          not null default ''
);

create trigger update_updated_at_timestamp
before update on certifications
for each row execute function update_updated_at_timestamp();

create index if not exists certifications_owner_id_idx on certifications (owner_id);

--------------------------------------------------------------------------------

