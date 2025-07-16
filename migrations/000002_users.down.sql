drop index if exists sessions_expiry_idx;
drop table if exists sessions;

--------------------------------------------------------------------------------

drop index if exists user_email_idx;
drop table if exists users;

--------------------------------------------------------------------------------

drop function if exists update_updated_at_timestamp restrict;

