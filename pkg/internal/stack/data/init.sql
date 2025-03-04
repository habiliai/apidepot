grant usage on schema public to anon, authenticated, service_role;
alter default privileges in schema public grant all on tables to anon, authenticated, service_role;
alter default privileges in schema public grant all on functions to anon, authenticated, service_role;
alter default privileges in schema public grant all on sequences to anon, authenticated, service_role;

CREATE SCHEMA IF NOT EXISTS stack;
CREATE TABLE IF NOT EXISTS stack.vapi_schema_migrations
(
    version         TIMESTAMP NOT NULL,
    vapi_package_id BIGINT    NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (version, vapi_package_id)
);
-- name 'migrations' is duplicated 'storage.migrations' for supabase/storage-api
CREATE TABLE IF NOT EXISTS stack.schema_migrations
(
    version    TIMESTAMP NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE SCHEMA IF NOT EXISTS api;
grant usage on schema api to anon, authenticated, service_role;
alter default privileges in schema api grant all on tables to anon, authenticated, service_role;
alter default privileges in schema api grant all on functions to anon, authenticated, service_role;
alter default privileges in schema api grant all on sequences to anon, authenticated, service_role;

CREATE SCHEMA IF NOT EXISTS storage;
grant usage on schema storage to anon, authenticated, service_role;
alter default privileges in schema storage grant all on tables to anon, authenticated, service_role;
alter default privileges in schema storage grant all on functions to anon, authenticated, service_role;
alter default privileges in schema storage grant all on sequences to anon, authenticated, service_role;

CREATE SCHEMA IF NOT EXISTS auth;
GRANT USAGE ON SCHEMA auth TO anon, authenticated, service_role;

create
    or replace function auth.uid()
    returns uuid
    language sql stable
as $$
select coalesce(
               current_setting('request.jwt.claim.sub', true),
               (current_setting('request.jwt.claims', true)::jsonb ->> 'sub')
       ) ::uuid
$$;

create
    or replace function auth.role()
    returns text
    language sql stable
as $$
select coalesce(
               current_setting('request.jwt.claim.role', true),
               (current_setting('request.jwt.claims', true)::jsonb ->> 'role')
       ) ::text
$$;

create
    or replace function auth.email()
    returns text
    language sql stable
as $$
select coalesce(
               current_setting('request.jwt.claim.email', true),
               (current_setting('request.jwt.claims', true)::jsonb ->> 'email')
       ) ::text
$$;
