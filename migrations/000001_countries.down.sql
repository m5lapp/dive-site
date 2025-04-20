set constraints all deferred;

--------------------------------------------------------------------------------

truncate table countries restart identity;
drop index if exists country_iso_number_idx;
drop index if exists country_iso2_code_idx;
drop index if exists country_iso3_code_idx;
drop table if exists countries;

--------------------------------------------------------------------------------

truncate table currencies restart identity;
drop index if exists currency_iso_alpha_idx;
drop table if exists currencies;

--------------------------------------------------------------------------------

set constraints all immediate;

