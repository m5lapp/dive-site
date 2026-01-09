create table if not exists entry_points (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into entry_points (sort, is_default, name, description) values
    (10, true,  'Boat', 'A dive from a boat'),
    (20, false, 'Pier/Jetty', 'A dive from a pier or jetty'),
    (30, false, 'Shore/Beach', 'A dive from the shore or a beach'),
    (40, false, 'Other', 'A different entry point');

--------------------------------------------------------------------------------

create table if not exists gas_mixes (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into gas_mixes (sort, is_default, name, description) values
    (10, true,  'Air',    'Normal air (21% oxygen, 79% nitrogen)'),
    (20, false, 'Heliox', 'Helium and oxygen'),
    (30, false, 'Nitrox', 'Enriched Air Nitrox (EAN)'),
    (40, false, 'Oxygen', 'Pure oxygen'),
    (50, false, 'Trimix', 'Helium, nitrogen and oxygen');

--------------------------------------------------------------------------------

create table if not exists currents (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into currents (sort, is_default, name, description) values
    (10, false, 'N/A', 'Not applicable due to the environment'),
    (20, false, 'None', 'There is no noticeable current'),
    (30, false, 'Light', 'A noticeable, but weak current'),
    (40, false, 'Moderate', 'A reasonably strong current'),
    (50, false, 'Strong', 'A powerful current'),
    (60, false, 'Very Strong', 'An extremely powerful current '),
    (70, false, 'Rip', 'A strong, localized and narrow current of water that moves directly away from the shore');

--------------------------------------------------------------------------------

create table if not exists waves (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into waves (sort, is_default, name, description) values
    (10, false, 'N/A', 'Not applicable due to the environment'),
    (20, false, 'Flat', 'The water is completely flat, like a mirror'),
    (30, false, 'Slight', 'Small wavelets, still short but more pronounced; crests have a glassy appearance and do not break'),
    (40, false, 'Moderate', 'Small waves with breaking crests, fairly frequent whitecap'),
    (50, false, 'Rough', 'Long waves begin to form, white foam crests are very frequent, some airborne spray is present'),
    (60, false, 'Very Rough', 'Foam from breaking waves is blown into streaks along the wind direction, considerable airborne spray is present');

--------------------------------------------------------------------------------

create table if not exists tank_configurations (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null,
    tank_count  smallint     not null default 0
);

insert into tank_configurations (sort, is_default, name, description, tank_count) values
    (10, false, 'Rebreather (CCR)', 'Closed circuit rebreather', 0),
    (20, false, 'Rebreather (SCR)', 'Semi-closed circuit rebreather', 0),
    (30, false, 'Sidemount', 'Two cylinders mounted on each side', 2),
    (40, true,  'Single Tank', 'One cylinder mounted on the back', 1),
    (50, false, 'Twinset', 'Twin cylinders mounted on the back', 2);

--------------------------------------------------------------------------------

create table if not exists tank_materials (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into tank_materials (sort, is_default, name, description) values
    (10, false, 'Aluminium', 'Commonly used in warmer waters'),
    (20, true,  'Steel', 'Commonly used in cooler waters');

--------------------------------------------------------------------------------

