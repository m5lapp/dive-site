set constraints all deferred;

--------------------------------------------------------------------------------

truncate table dive_sites restart identity;
drop index if exists dive_site_name_idx;
drop index if exists dive_site_country_idx;
drop table if exists dive_sites;

--------------------------------------------------------------------------------

truncate table water_bodies restart identity;
drop table if exists water_bodies;

--------------------------------------------------------------------------------

truncate table water_types restart identity;
drop table if exists water_types;

--------------------------------------------------------------------------------

set constraints all immediate;

