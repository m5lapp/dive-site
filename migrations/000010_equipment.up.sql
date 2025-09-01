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

