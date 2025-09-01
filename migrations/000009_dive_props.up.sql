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

