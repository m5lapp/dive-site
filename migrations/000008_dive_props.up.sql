create table if not exists dive_properties (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into dive_properties (sort, is_default, name, description) values
    ( 100, false, 'Cave Dive', 'A dive deep into a cave'),
    ( 200, false, 'Cavern Dive', 'A dive into a cave within the reach of natural light'),
    ( 400, false, 'Ice Dive', 'A dive under ice'),
    ( 600, false, 'Mine Dive', 'A dive in a flooded mine'),
    (1000, false, 'Reef Dive', 'A dive on a coral reef'),
    (1500, false, 'Wreck Dive (no pen)', 'Wreck dive with no-penetration'),
    (1200, false, 'Wall Dive', 'A dive along a vertical wall face'),
    ( 700, false, 'Muck Dive', 'A dive in the sediment that lies at the bottom of a dive site'),
    ( 500, false, 'Macro Dive', 'A dive with a focus on smaller marine creatures'),
    (1100, false, 'Shark Dive', 'A dive with a focus on seeing sharks'),
    (1300, false, 'Wreck Dive (limited pen)', 'A wreck dive remaining within sight of natural light'),
    (1400, false, 'Wreck Dive (full pen)', 'A wreck dive with full penetration of the wreck'),
    ( 300, false, 'Drift Dive', 'A dive drifting along with the current'),
    ( 800, false, 'Night Dive', 'A dive after dusk in total- or near-dark conditions'),
    ( 900, false, 'Pro Dive', 'A dive in the role of a paid dive professional');

--------------------------------------------------------------------------------

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

create table if not exists equipment (
    id          smallint     primary key generated always as identity,
    sort        smallint     not null unique,
    is_default  boolean      not null default false,
    name        varchar(32)  not null unique,
    description varchar(256) not null
);

insert into equipment (sort, is_default, name, description) values
    ( 10, false, 'Boots (3mm)', '3mm Wetsuit boots'),
    ( 20, false, 'Boots (5mm)', '5mm Wetsuit boots'),
    ( 30, false, 'Boots (7mm)', '7mm Wetsuit boots'),
    ( 40, false, 'Dry Suit', 'Dry suit'),
    ( 50, false, 'Gloves', 'Gloves'),
    ( 60, false, 'Hood (3mm)', '3mm Dive hood'),
    ( 70, false, 'Hood (5mm)', '5mm Dive hood'),
    ( 80, false, 'Hood (7mm)', '7mm Dive hood'),
    ( 90, false, 'Rash Vest', 'Rash vest'),
    (100, false, 'Skin', 'No exposure protection'),
    (110, false, 'Wetsuit (long, 3mm)', '3mm full-length wetsuit'),
    (120, false, 'Wetsuit (long, 5mm)', '5mm full-length wetsuit'),
    (130, false, 'Wetsuit (long, 7mm)', '7mm full-length wetsuit'),
    (140, false, 'Wetsuit (short, 3mm)', '3mm short wetsuit'),
    (150, false, 'Wetsuit (short, 5mm)', '5mm short wetsuit');

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
    description varchar(256) not null
);

insert into tank_configurations (sort, is_default, name, description) values
    (10, false, 'Rebreather (CCR)', 'Closed circuit rebreather'),
    (20, false, 'Rebreather (SCR)', 'Semi-closed circuit rebreather'),
    (30, false, 'Sidemount', 'Two cylinders mounted on each side '),
    (40, true,  'Single Tank', 'One cylinder mounted on the back'),
    (50, false, 'Twinset', 'Twin cylinders mounted on the back');

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

