create extension if not exists citext;

create table if not exists buddies (
    id                bigint       primary key generated always as identity,
    version           integer      not null default 1,
    created_at        timestamp(8) with time zone not null default now(),
    updated_at        timestamp(8) with time zone not null default now(),
    owner_id          bigint       not null references users(id) on delete cascade,
    buddy_user_id     bigint           null references users(id) on delete set null,
    name              varchar(256) not null,
    email             citext       not null default '',
    phone_number      varchar(32)  not null default '',
    agency_id         bigint           null references agencies(id) on delete set null,
    agency_member_num varchar(16)  not null default '',
    notes             text         not null default '',
    unique(owner_id, buddy_user_id)
);

create index if not exists buddies_owner_id_idx on buddies (owner_id);

--------------------------------------------------------------------------------

create table buddy_roles (
    id          smallint     primary key generated always as identity,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

--------------------------------------------------------------------------------

insert into buddy_roles (name, description) values
    ('Buddy', 'Dive buddy'),
    ('Customer', 'Paying customer'),
    ('Divemaster', 'Certified divemaster'),
    ('Divemaster Trainee', 'A divemaster in training'),
    ('Instructor', 'An instructor on a training course'),
    ('Instructor Trainer', 'As part of an instructor training program');

