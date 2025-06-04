create table if not exists trips (
    id          bigint        primary key generated always as identity,
    created_at  timestamp(8)  with time zone not null default now(),
    updated_at  timestamp(8)  with time zone not null default now(),
    owner_id    bigint        not null references users(id) on delete cascade,
    name        varchar(256)  not null,
    start_date  timestamp(8)  with time zone not null default now(),
    end_date    timestamp(8)  with time zone not null default now(),
    description varchar(1024) not null default '',
    rating      smallint,
    operator_id bigint                 references operators(id) on delete restrict,
    price       numeric(13, 3),
    currency_id smallint               references currencies(id) on delete restrict,
    notes       text          not null default ''
);

create index if not exists trips_owner_id_idx on trips (owner_id);

--------------------------------------------------------------------------------

create table if not exists courses (
    id            bigint        primary key generated always as identity,
    created_at    timestamp(8)  with time zone not null default now(),
    updated_at    timestamp(8)  with time zone not null default now(),
    owner_id      bigint        not null references users(id) on delete cascade,
    agency_course_id smallint            references agency_courses(id) on delete restrict,
    start_date    timestamp(8)  with time zone not null default now(),
    end_date      timestamp(8)  with time zone not null default now(),
    operator_id   bigint                 references operators(id) on delete restrict,
    instructor_id bigint                 references buddies(id) on delete restrict,
    price         numeric(13, 3),
    currency_id   smallint               references currencies(id) on delete restrict,
    rating        smallint,
    notes         text          not null default ''
);

create index if not exists courses_owner_id_idx on courses (owner_id);

--------------------------------------------------------------------------------

