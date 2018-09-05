package migration

import (
	"fmt"
	"strings"
)
var __000_bootstrap_sql = []byte(`-- This table sets the foundation for all future database migrations.
create table version (
  id bigserial primary key,
  updated_at timestamp with time zone not null default current_timestamp,
  version int not null unique CONSTRAINT positive_version CHECK (version >= 0)
);`)

func _000_bootstrap_sql() ([]byte, error) {
	return __000_bootstrap_sql, nil
}

var __001_common_sql = []byte(`-- tracker_items

CREATE TABLE tracker_items (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id bigserial primary key,
    remote_item_id text,
    item text,
    batch_id text,
    tracker_query_id bigint
);

-- trackers

CREATE TABLE trackers (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id bigserial primary key,
    url text,
    type text
);

-- tracker_queries

CREATE TABLE tracker_queries (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id bigserial primary key,
    query text,
    schedule text,
    tracker_id bigint
);

ALTER TABLE ONLY tracker_queries
    ADD CONSTRAINT tracker_queries_tracker_id_trackers_id_foreign 
        FOREIGN KEY (tracker_id)
        REFERENCES trackers(id)
        ON UPDATE RESTRICT 
        ON DELETE RESTRICT;

-- work_item_types

CREATE TABLE work_item_types (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text primary key,
    version integer,
    parent_path text,
    fields jsonb
);

-- work_items

CREATE TABLE work_items (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id bigserial primary key,
    type text,
    version integer,
    fields jsonb
);

`)

func _001_common_sql() ([]byte, error) {
	return __001_common_sql, nil
}

var __002_tracker_items_sql = []byte(`ALTER TABLE ONLY tracker_items
    ADD COLUMN tracker_id bigint;

ALTER TABLE ONLY tracker_items
    ADD CONSTRAINT tracker_items_tracker_id_trackers_id_foreign 
        FOREIGN KEY (tracker_id)
        REFERENCES trackers(id)
        ON UPDATE RESTRICT 
        ON DELETE RESTRICT;

ALTER TABLE ONLY tracker_items
    ADD CONSTRAINT tracker_items_remote_item_id_tracker_id_uni_idx
        UNIQUE (remote_item_id, tracker_id);
`)

func _002_tracker_items_sql() ([]byte, error) {
	return __002_tracker_items_sql, nil
}

var __003_login_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- login

CREATE TABLE identities (
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    full_name text,
    image_url text
);


-- user

CREATE TABLE users (
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text,
    identity_id uuid
);

CREATE UNIQUE INDEX uix_users_email ON users USING btree (email);`)

func _003_login_sql() ([]byte, error) {
	return __003_login_sql, nil
}

var __004_drop_tracker_query_id_sql = []byte(`ALTER TABLE tracker_items DROP COLUMN tracker_query_id;
`)

func _004_drop_tracker_query_id_sql() ([]byte, error) {
	return __004_drop_tracker_query_id_sql, nil
}

var __005_add_search_index_sql = []byte(`-- Add field on work_item  to store Full Text Search Vector
ALTER TABLE work_items ADD tsv tsvector;

UPDATE work_items SET tsv =
    setweight(to_tsvector('english', id::text),'A') ||
    setweight(to_tsvector('english', coalesce(fields->>'system.title','')),'B') ||
    setweight(to_tsvector('english', coalesce(fields->>'system.description','')),'C');

CREATE INDEX fulltext_search_index ON work_items USING GIN (tsv);

CREATE FUNCTION workitem_tsv_trigger() RETURNS trigger AS $$
begin
  new.tsv :=
    setweight(to_tsvector('english', new.id::text),'A') ||
    setweight(to_tsvector('english', coalesce(new.fields->>'system.title','')),'B') ||
    setweight(to_tsvector('english', coalesce(new.fields->>'system.description','')),'C');
  return new;
end
$$ LANGUAGE plpgsql;

CREATE TRIGGER upd_tsvector BEFORE INSERT OR UPDATE OF id, fields ON work_items
FOR EACH ROW EXECUTE PROCEDURE workitem_tsv_trigger();`)

func _005_add_search_index_sql() ([]byte, error) {
	return __005_add_search_index_sql, nil
}

var __006_rename_parent_path_sql = []byte(`ALTER TABLE work_item_types RENAME COLUMN "parent_path" TO "path";
UPDATE work_item_types set path=path || (case when path != '/' then '/'  else '' end) || name;`)

func _006_rename_parent_path_sql() ([]byte, error) {
	return __006_rename_parent_path_sql, nil
}

var __007_work_item_links_sql = []byte(`-- Here's the layout I'm trying to create:
-- (NOTE: work_items and work_item_types tables already exist)
--        
--           .----------------.
--           | work_items     |        .-----------------.
--           | ----------     |        | work_item_types |
--     .------>id bigserial   |        | --------------- |
--     |     | type text ------------>>| name text       |
--     |     | [other fields] |    |   | [other fields]  |
--     |     '----------------'    |   '-----------------'
--     |                           |
--     |   .------------------.    |   .-----------------------.
--     |   | work_item_links  |    |   | work_item_link_types  |
--     |   | ---------------  |    |   | ---------             |
--     |   | id uuid          | .------> id uuid               |
--     .-----source_id bigint | |  |   | name text             |
--      '----target_id bigint |/   |   | description text      |
--         | link_type_id uuid|    '-----source_type_name text |
--         '------------------'     '----target_type_name text |
--                                     | forward_name text     |
--                                     | reverse_name text     |
--    .--------------------------.    .- link_category_id uuid |
--    |work_item_link_categories |   / '-----------------------'
--    |------------------------- |  /
--    | id uuid                 <---
--    | name text                |
--    | description text         |
--    '--------------------------'

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- work item link categories

CREATE TABLE work_item_link_categories (
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone DEFAULT NULL,

    id          uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    version     integer DEFAULT 0 NOT NULL,

    name        text NOT NULL,
    description text
);

-- Ensure we only have one link category with the same name in existence.
-- If a category has been deleted (deleted_at != NULL) then we can recreate the category with the same name again.
CREATE UNIQUE INDEX work_item_link_categories_name_idx ON work_item_link_categories (name) WHERE deleted_at IS NULL;

-- work item link types

CREATE TYPE work_item_link_topology AS ENUM ('network', 'directed_network', 'dependency', 'tree');

CREATE TABLE work_item_link_types (
    created_at          timestamp with time zone,
    updated_at          timestamp with time zone,
    deleted_at          timestamp with time zone DEFAULT NULL,
    
    id                  uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    version             integer DEFAULT 0 NOT NULL,

    name                text NOT NULL,
    description         text,
    source_type_name    text REFERENCES work_item_types(name) ON DELETE CASCADE,
    target_type_name    text REFERENCES work_item_types(name) ON DELETE CASCADE,
    forward_name        text NOT NULL, -- MUST not be NULL because UI needs this
    reverse_name        text NOT NULL, -- MUST not be NULL because UI needs this
    topology            work_item_link_topology NOT NULL, 
    link_category_id    uuid REFERENCES work_item_link_categories(id) ON DELETE CASCADE
);

-- Ensure we only have one link type with the same name in a category in existence.
-- If a link type has been deleted (deleted_at != NULL) then we can recreate the link type with the same name again.
CREATE UNIQUE INDEX work_item_link_types_name_idx ON work_item_link_types (name, link_category_id) WHERE deleted_at IS NULL;

-- work item links

CREATE TABLE work_item_links (
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    deleted_at      timestamp with time zone DEFAULT NULL,
    
    id              uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    version         integer DEFAULT 0 NOT NULL,

    link_type_id    uuid REFERENCES work_item_link_types(id) ON DELETE CASCADE,
    source_id       bigint REFERENCES work_items(id) ON DELETE CASCADE,
    target_id       bigint REFERENCES work_items(id) ON DELETE CASCADE
);`)

func _007_work_item_links_sql() ([]byte, error) {
	return __007_work_item_links_sql, nil
}

var __008_soft_delete_or_resurrect_sql = []byte(`--#################################################################################
-- When a work item gets soft deleted, soft delete any work item link in existence.
--#################################################################################

CREATE FUNCTION update_WIL_after_WI() RETURNS trigger AS $update_WIL_after_WI$
    BEGIN
        UPDATE work_item_links SET deleted_at = NEW.deleted_at WHERE NEW.id IN (source_id, target_id);
        RETURN NEW;
    END;
$update_WIL_after_WI$ LANGUAGE plpgsql;

CREATE TRIGGER update_WIL_after_WI_trigger
AFTER UPDATE OF deleted_at
ON work_items
FOR EACH ROW
EXECUTE PROCEDURE update_WIL_after_WI();

--###########################################################################################
-- When a work item type gets soft deleted, soft delete any work item link type in existence.
--###########################################################################################

CREATE FUNCTION update_WILT_after_WIT() RETURNS trigger AS $update_WILT_after_WIT$
    BEGIN
        UPDATE work_item_link_types SET deleted_at = NEW.deleted_at WHERE NEW.name IN (source_type_name, target_type_name);
        RETURN NEW;
    END;
$update_WILT_after_WIT$ LANGUAGE plpgsql;

CREATE TRIGGER update_WILT_after_WIT_trigger
AFTER UPDATE OF deleted_at
ON work_item_types
FOR EACH ROW
EXECUTE PROCEDURE update_WILT_after_WIT();

--##################################################################################################
-- When a work item link category is soft deleted, soft delete any work item link type in existence.
--##################################################################################################

CREATE FUNCTION update_WILT_after_WILC() RETURNS trigger AS $update_WILT_after_WILC$
    BEGIN
        UPDATE work_item_link_types SET deleted_at = NEW.deleted_at WHERE link_category_id = NEW.id;
        RETURN NEW;
    END;
$update_WILT_after_WILC$ LANGUAGE plpgsql;

CREATE TRIGGER update_WILT_after_WILC_trigger
AFTER UPDATE OF deleted_at
ON work_item_link_categories
FOR EACH ROW
EXECUTE PROCEDURE update_WILT_after_WILC();

--##########################################################################################
-- When a work item link type is soft deleted, soft delete any work item links in existence.
--##########################################################################################

CREATE FUNCTION update_WIL_after_WILT() RETURNS trigger AS $update_WIL_after_WILT$
    BEGIN
        UPDATE work_item_links SET deleted_at = NEW.deleted_at WHERE link_type_id = NEW.id;
        RETURN NEW;
    END;
$update_WIL_after_WILT$ LANGUAGE plpgsql;

CREATE TRIGGER update_WIL_after_WILT_trigger
AFTER UPDATE OF deleted_at
ON work_item_link_types
FOR EACH ROW
EXECUTE PROCEDURE update_WIL_after_WILT();

--###############################################################################
-- When a work item type is soft deleted, soft delete any work item in existence.
--###############################################################################

CREATE FUNCTION update_WI_after_WIT() RETURNS trigger AS $update_WI_after_WIT$
    BEGIN
        UPDATE work_items SET deleted_at = NEW.deleted_at WHERE type = NEW.name;
        RETURN NEW;
    END;
$update_WI_after_WIT$ LANGUAGE plpgsql;

CREATE TRIGGER update_WI_after_WILT_trigger
AFTER UPDATE OF deleted_at
ON work_item_types
FOR EACH ROW
EXECUTE PROCEDURE update_WI_after_WIT();`)

func _008_soft_delete_or_resurrect_sql() ([]byte, error) {
	return __008_soft_delete_or_resurrect_sql, nil
}

var __009_drop_wit_trigger_sql = []byte(`-- See https://github.com/fabric8-services/fabric8-wit/issues/518 for an explanation
-- why these triggers were problematic.
DROP TRIGGER update_WI_after_WILT_trigger ON work_item_types;
DROP FUNCTION update_WI_after_WIT();
`)

func _009_drop_wit_trigger_sql() ([]byte, error) {
	return __009_drop_wit_trigger_sql, nil
}

var __010_comments_sql = []byte(`CREATE TABLE comments (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    parent_id text,
    body text,
    created_by uuid
);

CREATE INDEX ix_parent_id ON comments USING btree (parent_id);`)

func _010_comments_sql() ([]byte, error) {
	return __010_comments_sql, nil
}

var __011_projects_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- work item link categories

CREATE TABLE projects (
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone DEFAULT NULL,

    id          uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    version     integer DEFAULT 0 NOT NULL,

    name        text NOT NULL CHECK(name <> '')
);
CREATE UNIQUE INDEX projects_name_idx ON projects (name) WHERE deleted_at IS NULL;`)

func _011_projects_sql() ([]byte, error) {
	return __011_projects_sql, nil
}

var __012_unique_work_item_links_sql = []byte(`-- Delete duplicate links in existence and keep only one
-- See here: https://wiki.postgresql.org/wiki/Deleting_duplicates
DELETE FROM work_item_links
WHERE id IN (
    SELECT id
    FROM (
        SELECT id, ROW_NUMBER() OVER (partition BY link_type_id, source_id, target_id ORDER BY id) AS rnum
        FROM work_item_links
    ) t
    WHERE t.rnum > 1
);

-- From now on ensure we only have ONE link with the same source, target and
-- link type in existence. If a link has been deleted (deleted_at != NULL) then
-- we can recreate the link with the source, target and link type again.
CREATE UNIQUE INDEX work_item_links_unique_idx ON work_item_links (source_id, target_id, link_type_id) WHERE deleted_at IS NULL;`)

func _012_unique_work_item_links_sql() ([]byte, error) {
	return __012_unique_work_item_links_sql, nil
}

var __013_iterations_sql = []byte(`CREATE TABLE iterations (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    project_id uuid,
    parent_id uuid,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    name text
);

CREATE INDEX ix_project_id ON iterations USING btree (project_id);`)

func _013_iterations_sql() ([]byte, error) {
	return __013_iterations_sql, nil
}

var __014_wi_fields_index_sql = []byte(`CREATE INDEX work_items_fields_index on work_items USING gin(fields); `)

func _014_wi_fields_index_sql() ([]byte, error) {
	return __014_wi_fields_index_sql, nil
}

var __015_rename_projects_to_spaces_sql = []byte(`ALTER TABLE projects RENAME TO spaces;
ALTER INDEX projects_name_idx RENAME TO spaces_name_idx;
ALTER INDEX projects_pkey RENAME TO spaces_pkey;
ALTER TABLE spaces RENAME CONSTRAINT projects_name_check TO spaces_name_check;
ALTER TABLE iterations RENAME COLUMN project_id to space_id;`)

func _015_rename_projects_to_spaces_sql() ([]byte, error) {
	return __015_rename_projects_to_spaces_sql, nil
}

var __016_drop_wi_links_trigger_sql = []byte(`-- Dropping all auto soft-delete triggers

-- If a WIL is manually deleted, then recreated with same target/source/type, and a WIT is updated,
-- all WIL of that type is reset to have same state as the WIT and causing unique constraint problems.

-- A WIT can not be deleted, it will only be disabled from view to be created
-- A WI is only ever soft deleted, but a WIL to a deleted item should still be displayable.
-- A WITC delete is a process and can probably never be deleted without a large user question of; what do you want to do with these?
-- A WILT delete can never happen, or similar to above.

DROP TRIGGER update_WIL_after_WI_trigger ON work_items;
DROP FUNCTION update_WIL_after_WI();

DROP TRIGGER update_WILT_after_WIT_trigger ON work_item_types;
DROP FUNCTION update_WILT_after_WIT();

DROP TRIGGER update_WILT_after_WILC_trigger ON work_item_link_categories;
DROP FUNCTION update_WILT_after_WILC();

DROP TRIGGER update_WIL_after_WILT_trigger ON work_item_link_types;
DROP FUNCTION update_WIL_after_WILT();`)

func _016_drop_wi_links_trigger_sql() ([]byte, error) {
	return __016_drop_wi_links_trigger_sql, nil
}

var __017_alter_iterations_sql = []byte(`ALTER TABLE iterations ADD description TEXT;`)

func _017_alter_iterations_sql() ([]byte, error) {
	return __017_alter_iterations_sql, nil
}

var __018_rewrite_wits_sql = []byte(`-- See https://www.postgresql.org/docs/current/static/ltree.html for the
-- reference See http://leapfrogonline.io/articles/2015-05-21-postgres-ltree/
-- for an explanation
CREATE EXTENSION IF NOT EXISTS "ltree";

-- The following update needs to be done in order to get the WIT storage in a
-- good shape for it to be migrated to an ltree
UPDATE work_item_types SET
    -- Remove any leading '/' from the WIT's path.
    -- Remove any occurence of 'system.'.
    -- Replace '/' with '.' as the new path separator for use with ltree.
    -- Replace every non-C-LOCALE character with an underscore (the "." is an
    -- exception because it will be used by ltree)
    path =  regexp_replace(
                replace(replace(ltrim(path, '/'), 'system.', ''), '/', '.'),
                '[^a-zA-Z0-9_\.]',
                '_'
            )
    ;

-- Convert the path column from type text to ltree
ALTER TABLE work_item_types ALTER COLUMN path TYPE ltree USING path::ltree;

-- Add a constraint to the work item type name 
ALTER TABLE work_item_types ADD CONSTRAINT work_item_link_types_check_name_c_locale CHECK (name ~ '[a-zA-Z0-9_]');

-- Add indexes 
CREATE INDEX wit_path_gist_idx ON work_item_types USING GIST (path);
CREATE INDEX wit_path_idx ON work_item_types USING BTREE (path);


---------------------------------------------------------------------------
-- Update work items and work item link types that point to the work items.
---------------------------------------------------------------------------


-- Drop the foreign keys on the work item links types that reference the work
-- item type. We add them back once we've changed the names and it is safe to
-- add the keys back again.
ALTER TABLE work_item_link_types DROP CONSTRAINT work_item_link_types_source_type_name_fkey;
ALTER TABLE work_item_link_types DROP CONSTRAINT work_item_link_types_target_type_name_fkey;

UPDATE work_item_link_types
SET
    source_type_name = subpath(wit_source.path, -1, 1),
    target_type_name = subpath(wit_target.path, -1, 1)
FROM
    work_item_types AS wit_source,
    work_item_types AS wit_target
WHERE
    source_type_name = wit_source.name
    AND target_type_name = wit_target.name;

-- Update work item's type
UPDATE work_items
SET type = subpath(wit.path, -1, 1)
FROM work_item_types AS wit
WHERE type = wit.name;

-- Use the leaf of the path "tree" as the name of the work item type
UPDATE work_item_types SET name = subpath(path, -1, 1);

-- Add foreign keys back in
ALTER TABLE work_item_link_types
    ADD CONSTRAINT "work_item_link_types_source_type_name_fkey"
    FOREIGN KEY (source_type_name)
    REFERENCES work_item_types(name)
    ON DELETE CASCADE;

ALTER TABLE work_item_link_types
    ADD CONSTRAINT "work_item_link_types_target_type_name_fkey"
    FOREIGN KEY (target_type_name)
    REFERENCES work_item_types(name)
    ON DELETE CASCADE;
`)

func _018_rewrite_wits_sql() ([]byte, error) {
	return __018_rewrite_wits_sql, nil
}

var __019_add_state_iterations_sql = []byte(`CREATE TYPE iteration_state AS ENUM ('new', 'start', 'close');
ALTER TABLE iterations ADD state iteration_state DEFAULT 'new';
`)

func _019_add_state_iterations_sql() ([]byte, error) {
	return __019_add_state_iterations_sql, nil
}

var __020_work_item_description_update_search_index_sql = []byte(`-- migrate work items description
update work_items set fields=jsonb_set(fields, '{system.description}', 
  jsonb_build_object('content', fields->>'system.description', 'markup', 'plain'))
  where fields->>'system.description' is not null;


-- update support for Full Text Search Vector on work item description
DROP TRIGGER IF EXISTS upd_tsvector ON work_items;
DROP FUNCTION IF EXISTS workitem_tsv_trigger() CASCADE;
DROP INDEX IF EXISTS fulltext_search_index;
CREATE INDEX fulltext_search_index ON work_items USING GIN (tsv);

-- update the 'tsv' column with the text value of the existing 'content' 
-- element in the 'system.description' JSON document
UPDATE work_items SET tsv =
    setweight(to_tsvector('english', id::text),'A') ||
    setweight(to_tsvector('english', coalesce(fields->>'system.title','')),'B') ||
    setweight(to_tsvector('english', coalesce(fields#>>'{system.description, content}','')),'C');

-- fill the 'tsv' column with the text value of the created/modified 'content' 
-- element in the 'system.description' JSON document
CREATE FUNCTION workitem_tsv_trigger() RETURNS trigger AS $$
begin
  new.tsv :=
    setweight(to_tsvector('english', new.id::text),'A') ||
    setweight(to_tsvector('english', coalesce(new.fields->>'system.title','')),'B') ||
    setweight(to_tsvector('english', coalesce(new.fields#>>'{system.description, content}','')),'C');
  return new;
end
$$ LANGUAGE plpgsql; 

CREATE TRIGGER upd_tsvector BEFORE INSERT OR UPDATE OF id, fields ON work_items
FOR EACH ROW EXECUTE PROCEDURE workitem_tsv_trigger();`)

func _020_work_item_description_update_search_index_sql() ([]byte, error) {
	return __020_work_item_description_update_search_index_sql, nil
}

var __021_add_space_description_sql = []byte(`ALTER TABLE spaces ADD description text;`)

func _021_add_space_description_sql() ([]byte, error) {
	return __021_add_space_description_sql, nil
}

var __022_work_item_description_update_sql = []byte(`-- migrate work items description by replacing 'plain' with 'PlainText' in the 'markup' element of 'system.description' 
update work_items set fields=jsonb_set(fields, '{system.description, markup}', 
  to_jsonb('PlainText'::text)) where fields->>'system.description' is not null;

`)

func _022_work_item_description_update_sql() ([]byte, error) {
	return __022_work_item_description_update_sql, nil
}

var __023_comment_markup_sql = []byte(`-- add a 'markup' column in the 'comments' table
alter table comments add column markup text; `)

func _023_comment_markup_sql() ([]byte, error) {
	return __023_comment_markup_sql, nil
}

var __024_comment_markup_default_sql = []byte(`-- add a 'markup' column in the 'comments' table
update comments set markup = 'PlainText' where markup = NULL; `)

func _024_comment_markup_default_sql() ([]byte, error) {
	return __024_comment_markup_default_sql, nil
}

var __025_refactor_identities_users_sql = []byte(`-- Refactor Identities and Users tables.
ALTER TABLE identities
    DROP full_name,
    DROP image_url,
    ADD username text,
    ADD provider text,
    ADD user_id uuid;

ALTER TABLE users
    DROP identity_id,
    ADD full_name text,
    ADD image_url text,
    ADD bio text,
    ADD url text;`)

func _025_refactor_identities_users_sql() ([]byte, error) {
	return __025_refactor_identities_users_sql, nil
}

var __026_areas_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "ltree";

CREATE TABLE areas (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    space_id uuid,
    version integer DEFAULT 0 NOT NULL,
    path ltree,
    name text
);`)

func _026_areas_sql() ([]byte, error) {
	return __026_areas_sql, nil
}

var __027_areas_index_sql = []byte(`CREATE INDEX area_path_gist_index ON areas USING GIST (path);
`)

func _027_areas_index_sql() ([]byte, error) {
	return __027_areas_index_sql, nil
}

var __028_identity_provider_url_sql = []byte(`-- Refactor Identities: rename column 'provider' to 'provider_type'
ALTER TABLE identities RENAME COLUMN provider to provider_type;
-- the add a column to store the URL of the profile on the remote workitem system.
ALTER TABLE identities ADD profile_url text;
-- index to query identity by profile_url, which must be unique 
CREATE UNIQUE INDEX uix_identity_profileurl ON identities USING btree (profile_url);
-- index to query identity by user_id
CREATE INDEX uix_identity_userid ON identities USING btree (user_id);`)

func _028_identity_provider_url_sql() ([]byte, error) {
	return __028_identity_provider_url_sql, nil
}

var __029_identities_foreign_key_sql = []byte(`-- Refactor Identities: add a foreign key constraint
alter table identities add constraint identities_user_id_users_id_fk foreign key (user_id) REFERENCES users (id);
`)

func _029_identities_foreign_key_sql() ([]byte, error) {
	return __029_identities_foreign_key_sql, nil
}

var __030_indentities_unique_index_sql = []byte(`-- replace the unique index on `+"`"+`profile_url`+"`"+` with a check on 'DELETED_AT' to support soft deletes.
DROP INDEX uix_identity_profileurl;
CREATE UNIQUE INDEX uix_identity_profileurl ON identities USING btree (profile_url) WHERE deleted_at IS NULL;
`)

func _030_indentities_unique_index_sql() ([]byte, error) {
	return __030_indentities_unique_index_sql, nil
}

var __031_iterations_parent_path_ltree_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "ltree";

-- Rename parent_id column
ALTER TABLE iterations RENAME parent_id to path;

-- Need to convert the path column to text in order to
-- replace non-locale characters with an underscore
ALTER TABLE iterations ALTER path TYPE text USING path::text;

-- Need to update values of Iteration's' ParentID in order to migrate it to ltree
-- Replace every non-C-LOCALE character with an underscore
UPDATE iterations SET path = regexp_replace(path, '[^a-zA-Z0-9_\.]', '_', 'g');

-- Finally values in path are now in good shape for ltree and can be casted automatically to type ltree
-- Convert the parent column from type uuid to ltree
ALTER TABLE iterations ALTER path TYPE ltree USING path::ltree;

-- Enable full text search operaions using GIST index on path
CREATE INDEX iteration_path_gist_idx ON iterations USING GIST (path);
`)

func _031_iterations_parent_path_ltree_sql() ([]byte, error) {
	return __031_iterations_parent_path_ltree_sql, nil
}

var __032_add_foreign_key_space_id_sql = []byte(`-- Modify space_id column to be a foreign key to spaces ID table
ALTER TABLE iterations ADD CONSTRAINT iterations_space_id_spaces_id_fk FOREIGN KEY (space_id) REFERENCES spaces (id) ON DELETE CASCADE;
ALTER TABLE areas ADD CONSTRAINT areas_space_id_spaces_id_fk FOREIGN KEY (space_id) REFERENCES spaces (id) ON DELETE CASCADE;
`)

func _032_add_foreign_key_space_id_sql() ([]byte, error) {
	return __032_add_foreign_key_space_id_sql, nil
}

var __033_add_space_id_wilt_sql = []byte(`-- Alter the table work_item_link_types
INSERT INTO spaces (created_at, updated_at, id, name, description) VALUES (now(),now(), '{{index . 0}}','{{index . 1}}','{{index . 2}}');
ALTER TABLE work_item_link_types ADD space_id uuid DEFAULT '{{index . 0}}' NOT NULL;
-- Once we set the values to the default. We drop this default constraint
ALTER TABLE work_item_link_types ALTER space_id DROP DEFAULT;

ALTER TABLE work_item_link_types ADD FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX ix_space_id ON work_item_link_types USING btree (space_id);
`)

func _033_add_space_id_wilt_sql() ([]byte, error) {
	return __033_add_space_id_wilt_sql, nil
}

var __034_space_owner_sql = []byte(`ALTER TABLE spaces ADD owner_id uuid;`)

func _034_space_owner_sql() ([]byte, error) {
	return __034_space_owner_sql, nil
}

var __035_wit_to_use_uuid_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "ltree";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

------------------------------------------------------------------------------
-- Update the work item type table itself:
--
-- 0. In parallel to the current primary key ("name" column), we'll add a column
-- "id" that will become the new primary key later down the road.
--
-- 1. For all system-defined WITs, do not use the UUID as it was generated
-- during the migration instead use the one that is defined in the code.
--
-- 2. Add new "description" column and fill with the default value of 'This is
-- the description for the work item type "X".'.
--
-- 3. Update the "path" column of the WIT table to use the new UUID (with "-"
-- replaced by "_") instead of the "name" column.
--
-- 4. Drop the constraint that limits the "name" column to be contain only
-- C-LOCALE characters. This is a human readable free form field now.
--
-- 5. Finally, switch to "id" column to be our new primary key.
-------------------------------------------------------------------------------

ALTER TABLE work_item_types ADD COLUMN id uuid DEFAULT uuid_generate_v4() NOT NULL;

-- Use WIT IDs define in the code
UPDATE work_item_types SET id = '{{index . 0}}' WHERE name = 'planneritem';
UPDATE work_item_types SET id = '{{index . 1}}' WHERE name = 'userstory';
UPDATE work_item_types SET id = '{{index . 2}}' WHERE name = 'valueproposition';
UPDATE work_item_types SET id = '{{index . 3}}' WHERE name = 'fundamental';
UPDATE work_item_types SET id = '{{index . 4}}' WHERE name = 'experience';
UPDATE work_item_types SET id = '{{index . 5}}' WHERE name = 'feature';
UPDATE work_item_types SET id = '{{index . 6}}' WHERE name = 'scenario';
UPDATE work_item_types SET id = '{{index . 7}}' WHERE name = 'bug';

ALTER TABLE work_item_types ADD COLUMN description text;
UPDATE work_item_types SET description = concat('This is the description for the work item type "', name, '".');

CREATE OR REPLACE FUNCTION UUIDToLtreeNode(u uuid, OUT node ltree) AS $$ BEGIN
-- Converts a UUID value into a value usable inside an Ltree 
    SELECT replace(u::text, '-', '_') INTO node;
END; $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION LtreeNodeToUUID(node ltree, OUT u uuid) AS $$ BEGIN
-- Converts an Ltree node into a UUID value 
    SELECT replace(node::text, '_', '-') INTO u;
END; $$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_new_wit_path(oldPath ltree, OUT newPath ltree) AS $$
-- Converts the oldPath ltree value which was planneritem.bug and so forth into
-- an ltree that is based on the UUID of a work item type.
    DECLARE
        nodeName text;
        nodeId text;
        newPathArray text array;
    BEGIN
        FOREACH nodeName IN array regexp_split_to_array(oldPath::text,'\.')
        LOOP
            SELECT UUIDToLtreeNode(id) INTO nodeId FROM work_item_types WHERE name = nodeName;
            newPathArray := array_append(newPathArray, nodeId);
        END LOOP;
        newPath := array_to_string(newPathArray, '.');
    END;
$$ LANGUAGE plpgsql;

UPDATE work_item_types SET path = get_new_wit_path(path);

DROP FUNCTION get_new_wit_path(oldPath ltree, OUT newPath ltree);

-- Drop constraints that depend on the primary key.
ALTER TABLE work_item_link_types DROP CONSTRAINT work_item_link_types_source_type_name_fkey;
ALTER TABLE work_item_link_types DROP CONSTRAINT work_item_link_types_target_type_name_fkey;
-- Drop the primary key itself and set up the new one on the "id" column.
ALTER TABLE work_item_types DROP CONSTRAINT work_item_types_pkey;
ALTER TABLE work_item_types ADD PRIMARY KEY (id);
ALTER TABLE work_item_types DROP CONSTRAINT work_item_link_types_check_name_c_locale;

------------------------------------------------------------------------------
-- Update all references to the work item type table to point to the new "id"
-- column instead of the "name" column. Since this involves column type change
-- from "text" to "uuid" we'll simply add a new reference and delete the old
-- one.
------------------------------------------------------------------------------

------------------------------
-- Update work item link types
------------------------------

ALTER TABLE work_item_link_types ADD COLUMN source_type_id uuid REFERENCES work_item_types(id) ON DELETE CASCADE;
ALTER TABLE work_item_link_types ADD COLUMN target_type_id uuid REFERENCES work_item_types(id) ON DELETE CASCADE;

UPDATE work_item_link_types SET source_type_id = (SELECT id FROM work_item_types WHERE name = source_type_name);
UPDATE work_item_link_types SET target_type_id = (SELECT id FROM work_item_types WHERE name = target_type_name);

ALTER TABLE work_item_link_types DROP COLUMN source_type_name;
ALTER TABLE work_item_link_types DROP COLUMN target_type_name;

-- Add NOT NULL criteria 
ALTER TABLE work_item_link_types ALTER COLUMN source_type_id SET NOT NULL;
ALTER TABLE work_item_link_types ALTER COLUMN target_type_id SET NOT NULL;

--------------------
-- Update work items
--------------------

-- NOTE: The foreign key is new!
ALTER TABLE work_items RENAME type TO type_old;
ALTER TABLE work_items ADD COLUMN type uuid REFERENCES work_item_types(id) ON DELETE CASCADE;
UPDATE work_items SET type = (SELECT id FROM work_item_types WHERE name = type_old);
ALTER TABLE work_items DROP COLUMN type_old;

-- Add NOT NULL criteria
ALTER TABLE work_items ALTER COLUMN type SET NOT NULL;`)

func _035_wit_to_use_uuid_sql() ([]byte, error) {
	return __035_wit_to_use_uuid_sql, nil
}

var __036_add_icon_to_wit_sql = []byte(`ALTER TABLE work_item_types ADD COLUMN icon text DEFAULT 'fa-question' NOT NULL;`)

func _036_add_icon_to_wit_sql() ([]byte, error) {
	return __036_add_icon_to_wit_sql, nil
}

var __037_work_item_revisions_sql = []byte(`-- create a revision table for work items, using the some columns + identity of the user and timestamp of the operation
CREATE TABLE work_item_revisions (
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    revision_time timestamp with time zone default current_timestamp,
    revision_type int NOT NULL,
    modifier_id uuid NOT NULL,
    work_item_id bigint NOT NULL,
    work_item_type_id uuid,
    work_item_version integer,
    work_item_fields jsonb
);

CREATE INDEX work_item_revisions_work_items_idx ON work_item_revisions USING BTREE (work_item_id);

ALTER TABLE work_item_revisions
    ADD CONSTRAINT work_item_revisions_identity_fk FOREIGN KEY (modifier_id) REFERENCES identities(id);

-- delete work item revisions when the work item is deleted from the database.
ALTER TABLE work_item_revisions
    ADD CONSTRAINT work_item_revisions_work_items_fk FOREIGN KEY (work_item_id) REFERENCES work_items(id) ON DELETE CASCADE;




`)

func _037_work_item_revisions_sql() ([]byte, error) {
	return __037_work_item_revisions_sql, nil
}

var __038_comment_revisions_sql = []byte(`-- create a revision table for comments, using the some columns + identity of the user and timestamp of the operation
CREATE TABLE comment_revisions (
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    revision_time timestamp with time zone default current_timestamp,
    revision_type int NOT NULL,
    modifier_id uuid NOT NULL,
    comment_id uuid NOT NULL,
    comment_body text,
    comment_markup text
);

CREATE INDEX comment_revisions_comment_id_idx ON comment_revisions USING BTREE (comment_id);

ALTER TABLE comment_revisions
    ADD CONSTRAINT comment_revisions_identity_fk FOREIGN KEY (modifier_id) REFERENCES identities(id);

-- delete comment revisions when the comment is deleted from the database.
ALTER TABLE comment_revisions
    ADD CONSTRAINT comment_revisions_comments_fk FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE;




`)

func _038_comment_revisions_sql() ([]byte, error) {
	return __038_comment_revisions_sql, nil
}

var __039_comment_revisions_parentid_sql = []byte(`-- add a 'comment_parent_id' column in the 'comment_revisions' table
ALTER TABLE comment_revisions ADD COLUMN comment_parent_id text;
-- fill the new column 
update comment_revisions set comment_parent_id = c.parent_id from comments c where c.id = comment_id;
-- make the new column 'not null'
ALTER TABLE comment_revisions ALTER COLUMN comment_parent_id SET NOT NULL;
-- make sure the new column cannot be filled with empty content
ALTER TABLE comment_revisions ADD CONSTRAINT comment_parent_id_check CHECK (comment_parent_id <> '');`)

func _039_comment_revisions_parentid_sql() ([]byte, error) {
	return __039_comment_revisions_parentid_sql, nil
}

var __040_add_space_id_wi_wit_tq_sql = []byte(`-- Alter the table work_items
ALTER TABLE work_items ADD space_id uuid DEFAULT '{{index . 0}}' NOT NULL;
-- Once we set the values to the default. We drop this default constraint
ALTER TABLE work_items ALTER space_id DROP DEFAULT;
ALTER TABLE work_items ADD FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX ix_work_items_space_id ON work_items USING btree (space_id);

-- Alter the table work_item_types
ALTER TABLE work_item_types ADD space_id uuid DEFAULT '{{index . 0}}' NOT NULL;
-- Once we set the values to the default. We drop this default constraint
ALTER TABLE work_item_types ALTER space_id DROP DEFAULT;
ALTER TABLE work_item_types ADD FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX ix_work_item_types_space_id ON work_item_types USING btree (space_id);


-- Alter the table tracker_queries
ALTER TABLE tracker_queries ADD space_id uuid DEFAULT '{{index . 0}}' NOT NULL;
-- Once we set the values to the default. We drop this default constraint
ALTER TABLE tracker_queries ALTER space_id DROP DEFAULT;
ALTER TABLE tracker_queries ADD FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;

-- Create indexes
CREATE INDEX ix_tracker_queries_space_id ON tracker_queries USING btree (space_id);
`)

func _040_add_space_id_wi_wit_tq_sql() ([]byte, error) {
	return __040_add_space_id_wi_wit_tq_sql, nil
}

var __041_unique_area_name_create_new_area_sql = []byte(`------ You can't allow the same area name and the same ancestry inside a space

ALTER TABLE areas ADD CONSTRAINT areas_name_space_id_path_unique UNIQUE(space_id,name,path);

------  For existing spaces in production, which dont have a default area, create one.
--
-- 1. Get all spaces which have an area under it with the same name.
-- 2. Get all spaces not in (1)
-- 3. insert an 'area' for all such spaces in (2)

INSERT INTO areas
            (created_at,
             updated_at,
             name,
             space_id)
SELECT current_timestamp,
       current_timestamp,
       name,
       id
FROM   spaces
WHERE  id NOT IN (SELECT s.id
                  FROM   spaces AS s
                         INNER JOIN areas AS a
                                 ON s.name = a.name
                                    AND s.id = a.space_id);  


----- for all other existing areas in production, move them under the default 'root' area.



                           
CREATE OR REPLACE FUNCTION GetRootArea(s_id uuid,OUT root_id uuid) AS $$ BEGIN
-- Get Root area for a space
     select id from areas 
          where name in ( SELECT name as space_name
          from spaces 
              where id=s_id )
                  and space_id =s_id 
                           into root_id;
END; $$ LANGUAGE plpgsql ;

-- Convert Text to Ltree , use standard library FUNCTION?

CREATE OR REPLACE FUNCTION TextToLtreeNode(u text, OUT node ltree) AS $$ BEGIN
    SELECT replace(u, '-', '_') INTO node;
END; $$ LANGUAGE plpgsql;



-- Migrate all existing areas into the new tree where the parent is always the root area

CREATE OR REPLACE FUNCTION GetUpdatedAreaPath(area_id uuid,space_id uuid, path ltree, OUT updated_path ltree) AS $rootarea$ 
-- Migrate all existing areas into the new tree where the parent is always the root area

     DECLARE 
          rootarea uuid;                                            
     BEGIN
     
     select GetRootArea(space_id) into rootarea;
     IF rootarea != area_id 
         THEN                  
         IF path=''
            THEN 
             select UUIDToLtreeNode(rootarea) into updated_path ;
         ELSE 
             select TextToLtreeNode(concat(rootarea::text,'.',path::text)) into updated_path;
         END IF;
     ELSE 
         updated_path:='';        
     END IF;
END; 
$rootarea$  LANGUAGE plpgsql ;   

-- Move all areas under that space into the root area ( except of course the root area ;) ), 

UPDATE AREAS set path=GetUpdatedAreaPath(id,space_id,path) ;

-- cleanup

DROP FUNCTION GetUpdatedAreaPath(uuid,uuid,ltree);
DROP FUNCTION GetRootArea(uuid);
DROP FUNCTION TextToLtreeNode(text);
`)

func _041_unique_area_name_create_new_area_sql() ([]byte, error) {
	return __041_unique_area_name_create_new_area_sql, nil
}

var __042_work_item_link_revisions_sql = []byte(`-- create a revision table for work item links, using the some columns + identity of the user and timestamp of the operation
CREATE TABLE work_item_link_revisions (
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    revision_time timestamp with time zone default current_timestamp,
    revision_type int NOT NULL,
    modifier_id uuid NOT NULL,
    work_item_link_id uuid NOT NULL,
    work_item_link_version int NOT NULL,
    work_item_link_source_id bigint NOT NULL,
    work_item_link_target_id bigint NOT NULL,
    work_item_link_type_id uuid NOT NULL
);

CREATE INDEX work_item_link_revisions_work_item_link_id_idx ON work_item_link_revisions USING BTREE (work_item_link_id);

ALTER TABLE work_item_link_revisions
    ADD CONSTRAINT work_item_link_revisions_modifier_id_fk FOREIGN KEY (modifier_id) REFERENCES identities(id);

-- delete work item revisions when the work item is deleted from the database.
ALTER TABLE work_item_link_revisions
    ADD CONSTRAINT work_item_link_revisions_work_item_link_id_fk FOREIGN KEY (work_item_link_id) REFERENCES work_item_links(id) ON DELETE CASCADE;

`)

func _042_work_item_link_revisions_sql() ([]byte, error) {
	return __042_work_item_link_revisions_sql, nil
}

var __043_space_resources_sql = []byte(`-- Create space resource table for Keycloak resources associated with spaces
CREATE TABLE space_resources (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    space_id uuid NOT NULL,
    resource_id text NOT NULL,
    policy_id text NOT NULL,
    permission_id text NOT NULL
);

CREATE INDEX space_resources_space_id_idx ON space_resources USING BTREE (space_id);

ALTER TABLE space_resources
    ADD CONSTRAINT space_resources_space_fk FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;
`)

func _043_space_resources_sql() ([]byte, error) {
	return __043_space_resources_sql, nil
}

var __044_add_contextinfo_column_users_sql = []byte(` alter table users add  context_information jsonb;
 `)

func _044_add_contextinfo_column_users_sql() ([]byte, error) {
	return __044_add_contextinfo_column_users_sql, nil
}

var __045_adds_order_to_existing_wi_sql = []byte(`ALTER TABLE work_items ADD COLUMN execution_order double precision;
CREATE INDEX order_index ON work_items (execution_order);

CREATE OR REPLACE FUNCTION adds_order() RETURNS void as $$
-- adds_order() function adds order to existing work_items in database
	DECLARE 
		i integer=1000;
		r RECORD;
		xyz CURSOR FOR SELECT id, execution_order from work_items;
	BEGIN
		open xyz;
			FOR r in FETCH ALL FROM xyz LOOP
				UPDATE work_items set execution_order=i where id=r.id;
				i := i+1000;
			END LOOP;
		close xyz;
END $$ LANGUAGE plpgsql;

DO $$ BEGIN
	PERFORM adds_order();
	DROP FUNCTION adds_order();
END $$;
`)

func _045_adds_order_to_existing_wi_sql() ([]byte, error) {
	return __045_adds_order_to_existing_wi_sql, nil
}

var __046_oauth_states_sql = []byte(`-- Create Oauth state reference table for states used in oauth workflow
CREATE TABLE oauth_state_references (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    referrer text NOT NULL
);`)

func _046_oauth_states_sql() ([]byte, error) {
	return __046_oauth_states_sql, nil
}

var __047_codebases_sql = []byte(`CREATE TABLE codebases (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    space_id uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    type text,
    url text
);

CREATE INDEX ix_codebases_space_id ON codebases USING btree (space_id);`)

func _047_codebases_sql() ([]byte, error) {
	return __047_codebases_sql, nil
}

var __048_unique_iteration_name_create_new_iteration_sql = []byte(`-- Update iterations having same name, append UUID to make it unique.
UPDATE iterations
SET name = name || '-' || uuid_generate_v4()
WHERE id IN
    (SELECT id
     FROM iterations
     WHERE name IN
         (SELECT name
          FROM iterations
          GROUP BY name
          HAVING count(name) >1));

------  For existing spaces in production, which dont have a root iteration, create one.
--
-- 1. Get all spaces which have an iteration under it with the same name.
-- 2. Get all spaces not in (1)
-- 3. insert an 'iteration' for all such spaces in (2)
INSERT INTO iterations
            (created_at,
             updated_at,
             name,
             space_id)
SELECT current_timestamp,
       current_timestamp,
       name,
       id
FROM   spaces
WHERE  id NOT IN (SELECT s.id
                  FROM   spaces AS s
                         INNER JOIN iterations AS i
                                 ON s.name = i.name
                                    AND s.id = i.space_id);

----- for all other existing iterations in production, move them under the root iteration of given space.
CREATE OR REPLACE FUNCTION GetRootIteration(s_id uuid,OUT root_id uuid) AS $$ BEGIN
-- Get Root iteration for a space
     select id from iterations 
          where name in ( SELECT name as space_name
          from spaces 
              where id=s_id )
                  and space_id =s_id 
                           into root_id;
END; $$ LANGUAGE plpgsql ;

-- Convert Text to Ltree , use standard library FUNCTION?

CREATE OR REPLACE FUNCTION TextToLtreeNode(u text, OUT node ltree) AS $$ BEGIN
    SELECT replace(u, '-', '_') INTO node;
END; $$ LANGUAGE plpgsql;


-- Migrate all existing iterations into the new tree where the parent is always the root iteration

CREATE OR REPLACE FUNCTION GetUpdatedIterationPath(iteration_id uuid,space_id uuid, path ltree, OUT updated_path ltree) AS $rootiteration$ 
-- Migrate all existing iterations into the new tree where the parent is always the root iteration

     DECLARE
          rootiteration uuid;
     BEGIN
     -- In production this probably not NULL; safety check.
     If path IS NULL
        THEN
            path = '';
     END IF;
     select GetRootIteration(space_id) into rootiteration;
     IF rootiteration != iteration_id 
         THEN                  
         IF path=''
            THEN 
             select UUIDToLtreeNode(rootiteration) into updated_path ;
         ELSE 
             select TextToLtreeNode(concat(rootiteration::text,'.',path::text)) into updated_path;
         END IF;
     ELSE 
         updated_path:='';        
     END IF;
END;
$rootiteration$  LANGUAGE plpgsql ;

-- Move all iterations under it's space and into the root iterations ( except of course the root iteration ;) ), 

UPDATE iterations set path=GetUpdatedIterationPath(id,space_id,path);

update work_items set fields=jsonb_set(fields, '{system.iteration}', to_jsonb(subq.id::text)) 
    from (select id, space_id from iterations where path = '') AS subq
    where subq.space_id = work_items.space_id and fields->>'system.iteration' IS NULL;

-- cleanup
DROP FUNCTION GetUpdatedIterationPath(uuid,uuid,ltree);
DROP FUNCTION GetRootIteration(uuid);
DROP FUNCTION TextToLtreeNode(text);

CREATE INDEX ix_name ON iterations USING btree (name);

------ You can't allow the same iteration name and the same ancestry inside a space
ALTER TABLE iterations ADD CONSTRAINT iterations_name_space_id_path_unique UNIQUE(space_id,name,path);
`)

func _048_unique_iteration_name_create_new_iteration_sql() ([]byte, error) {
	return __048_unique_iteration_name_create_new_iteration_sql, nil
}

var __049_add_wi_to_root_area_sql = []byte(`update work_items set fields=jsonb_set(fields, '{system.area}', to_jsonb(subq.id::text)) 
    from (select id, space_id from areas where path = '') AS subq
    where subq.space_id = work_items.space_id and fields->>'system.area' IS NULL;


CREATE INDEX ix_area_name ON areas USING btree (name);
`)

func _049_add_wi_to_root_area_sql() ([]byte, error) {
	return __049_add_wi_to_root_area_sql, nil
}

var __050_add_company_to_user_profile_sql = []byte(`ALTER TABLE users ADD COLUMN company TEXT;
`)

func _050_add_company_to_user_profile_sql() ([]byte, error) {
	return __050_add_company_to_user_profile_sql, nil
}

var __051_modify_work_item_link_types_name_idx_sql = []byte(`-- When created, a work item link type didn't know about any space and thus its
-- name was only allowed to be used once per link category. Now, with spaces,
-- the same unique index shall span the name, the category and the space.

DROP INDEX work_item_link_types_name_idx;

CREATE UNIQUE INDEX work_item_link_types_name_idx ON work_item_link_types (name, space_id, link_category_id) WHERE deleted_at IS NULL;`)

func _051_modify_work_item_link_types_name_idx_sql() ([]byte, error) {
	return __051_modify_work_item_link_types_name_idx_sql, nil
}

var __052_unique_space_names_sql = []byte(`-- drop existing unique index
DROP INDEX spaces_name_idx;
-- recreate unique index with original index name, on two columns
CREATE UNIQUE INDEX spaces_name_idx ON spaces (name, owner_id) WHERE deleted_at IS NULL;`)

func _052_unique_space_names_sql() ([]byte, error) {
	return __052_unique_space_names_sql, nil
}

var __053_edit_username_sql = []byte(`-- default is 'false', works with business logic as well.
ALTER TABLE identities ADD COLUMN registration_completed BOOLEAN NOT NULL DEFAULT FALSE;
`)

func _053_edit_username_sql() ([]byte, error) {
	return __053_edit_username_sql, nil
}

var __054_add_stackid_to_codebase_sql = []byte(`ALTER TABLE codebases ADD COLUMN stack_id TEXT;

-- Should we set the default to current codebases entries, hardcoded value is java-centos
UPDATE codebases set stack_id ='java-centos';
`)

func _054_add_stackid_to_codebase_sql() ([]byte, error) {
	return __054_add_stackid_to_codebase_sql, nil
}

var __055_assign_root_area_if_missing_sql = []byte(`update work_items set fields=jsonb_set(fields, '{system.area}', to_jsonb(subq.id::text)) 
    from (select id, space_id from areas where path = '') AS subq
    where subq.space_id = work_items.space_id and fields->>'system.area' IS NULL;`)

func _055_assign_root_area_if_missing_sql() ([]byte, error) {
	return __055_assign_root_area_if_missing_sql, nil
}

var __056_assign_root_iteration_if_missing_sql = []byte(`update work_items set fields=jsonb_set(fields, '{system.iteration}', to_jsonb(subq.id::text)) 
    from (select id, space_id from iterations where path = '') AS subq
    where subq.space_id = work_items.space_id and fields->>'system.iteration' IS NULL;`)

func _056_assign_root_iteration_if_missing_sql() ([]byte, error) {
	return __056_assign_root_iteration_if_missing_sql, nil
}

var __057_add_last_used_workspace_to_codebase_sql = []byte(`ALTER TABLE codebases ADD COLUMN last_used_workspace TEXT;
`)

func _057_add_last_used_workspace_to_codebase_sql() ([]byte, error) {
	return __057_add_last_used_workspace_to_codebase_sql, nil
}

var __058_index_identities_fullname_sql = []byte(`create index idx_user_full_name on users (lower(full_name));
create index idx_user_email on users (lower(email));
create index idx_idenities_username on identities (lower(username));`)

func _058_index_identities_fullname_sql() ([]byte, error) {
	return __058_index_identities_fullname_sql, nil
}

var __059_fixed_ids_for_system_link_types_and_categories_sql = []byte(`-- 1. First make sure that the link type IDs are consistent across all systems.
--    We therefore copy the existing link types and only change the ID value to
--    the ones pre-defined in code.
-- 2. Then we make sure all existing links that pointed to the old ID now point
--    to the duplicated row with the new ID.
-- 3. Then we can delete all the old link types.

-- The same we do for the "user" and "system" link categories.

--------------------
-- HANDLE LINK TYPES
--------------------

-- Duplicate the old SystemWorkItemLinkTypeBugBlocker and use a new ID
-- We have to use a different name in order to not violate the uniq key
-- "work_item_link_types_name_idx". The name will later be updated through
-- the migration.
INSERT INTO work_item_link_types(id, created_at, updated_at, deleted_at,
    name, version, topology, forward_name, reverse_name, link_category_id,
    source_type_id, target_type_id, space_id)
    SELECT '{{index . 0}}', created_at, updated_at, deleted_at,
    '{{index . 0}}', version, topology, forward_name, reverse_name, link_category_id,
    source_type_id, target_type_id, space_id 
    FROM work_item_link_types
    WHERE name='Bug blocker';

-- Duplicate the old SystemWorkItemLinkPlannerItemRelated and use a new ID
INSERT INTO work_item_link_types(id, created_at, updated_at, deleted_at,
    name, version, topology, forward_name, reverse_name, link_category_id,
    source_type_id, target_type_id, space_id)
    SELECT '{{index . 1}}', created_at, updated_at, deleted_at,
    '{{index . 1}}', version, topology, forward_name, reverse_name, link_category_id,
    source_type_id, target_type_id, space_id 
    FROM work_item_link_types
    WHERE name='Related planner item';

INSERT INTO work_item_link_types(id, created_at, updated_at, deleted_at,
    name, version, topology, forward_name, reverse_name, link_category_id,
    source_type_id, target_type_id, space_id)
    SELECT '{{index . 2}}', created_at, updated_at, deleted_at,
    '{{index . 2}}', version, topology, forward_name, reverse_name, link_category_id,
    source_type_id, target_type_id, space_id 
    FROM work_item_link_types
    WHERE name='Parent child item';

-- Update existing links to use the new link type ID
UPDATE work_item_links SET link_type_id='{{index . 0}}'
    WHERE link_type_id = (SELECT id FROM work_item_link_types WHERE name='Bug blocker' AND id <> '{{index . 0}}');

UPDATE work_item_links SET link_type_id='{{index . 1}}'
    WHERE link_type_id = (SELECT id FROM work_item_link_types WHERE name='Related planner item' AND id <> '{{index . 1}}');
    
UPDATE work_item_links SET link_type_id='{{index . 2}}'
    WHERE link_type_id = (SELECT id FROM work_item_link_types WHERE name='Parent child item' AND id <> '{{index . 2}}');

-- Delete old link types and only leave the new ones.
DELETE FROM work_item_link_types WHERE id NOT IN ('{{index . 0}}', '{{index . 1}}', '{{index . 2}}');

--------------------
-- HANDLE CATEGORIES
--------------------

-- Duplicate "system" link category
INSERT INTO work_item_link_categories (id, created_at, updated_at, deleted_at, name, version, description)
    SELECT '{{index . 3}}', created_at, updated_at, deleted_at, '{{index . 3}}', version, description
    FROM work_item_link_categories
    WHERE name='system';

-- Duplicate "user" link category
INSERT INTO work_item_link_categories (id, created_at, updated_at, deleted_at, name, version, description)
    SELECT '{{index . 4}}', created_at, updated_at, deleted_at, '{{index . 4}}', version, description
    FROM work_item_link_categories
    WHERE name='user';

-- Update existing link types to use the new category IDs
UPDATE work_item_link_types SET link_category_id='{{index . 3}}'
    WHERE link_category_id = (SELECT id FROM work_item_link_categories WHERE name='system' AND id <> '{{index . 3}}');

UPDATE work_item_link_types SET link_category_id='{{index . 4}}'
    WHERE link_category_id = (SELECT id FROM work_item_link_categories WHERE name='user' AND id <> '{{index . 4}}');

-- Delete old link categories and only leave the new ones.
DELETE FROM work_item_link_categories WHERE id NOT IN ('{{index . 3}}', '{{index . 4}}');
`)

func _059_fixed_ids_for_system_link_types_and_categories_sql() ([]byte, error) {
	return __059_fixed_ids_for_system_link_types_and_categories_sql, nil
}

var __060_fixed_identities_username_idx_sql = []byte(`-- drop existing unique index
DROP INDEX idx_idenities_username;
-- recreate unique index idx_identities_username case insensitive username
CREATE INDEX idx_identities_username ON identities (username);
`)

func _060_fixed_identities_username_idx_sql() ([]byte, error) {
	return __060_fixed_identities_username_idx_sql, nil
}

var __061_replace_index_space_name_sql = []byte(`-- drop existing unique index
DROP INDEX spaces_name_idx;
-- rename duplicate spaces in existence and keep only one as it was
UPDATE spaces SET name = CONCAT(lower(name), '-renamed')
WHERE id IN (
    SELECT id
    FROM (
        SELECT id, ROW_NUMBER() OVER (partition BY owner_id, lower(name) ORDER BY id) AS rnum
        FROM spaces
    ) t
    WHERE t.rnum > 1
);
-- recreate unique index with original index and lowercase name, on two columns
CREATE UNIQUE INDEX spaces_name_idx ON spaces (owner_id, lower(name)) WHERE deleted_at IS NULL;
`)

func _061_replace_index_space_name_sql() ([]byte, error) {
	return __061_replace_index_space_name_sql, nil
}

var __062_link_system_preparation_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE work_item_link_type_combinations (
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    deleted_at      timestamp with time zone,
    id uuid         primary key DEFAULT uuid_generate_v4() NOT NULL,
    version         integer,
    link_type_id    uuid NOT NULL REFERENCES work_item_link_types(id) ON DELETE CASCADE,
    source_type_id  uuid NOT NULL REFERENCES work_item_types(id) ON DELETE CASCADE,
    target_type_id  uuid NOT NULL REFERENCES work_item_types(id) ON DELETE CASCADE,
    -- We need the space id here because different space templates might specify
    -- the same source/target type combination for the same system-defined link
    -- type (e.g. "parent of"). That would violated our unique constraint below
    -- if the space_id was missing from it.
    -- TODO(kwk): once we have space templates, this will become a reference to
    -- the space template and not to the space.
    space_id        uuid  NOT NULL REFERENCES spaces(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX work_item_link_type_combinations_uniq
    ON work_item_link_type_combinations (
        space_id,
        link_type_id,
        source_type_id,
        target_type_id
    )
    WHERE deleted_at IS NULL;
`)

func _062_link_system_preparation_sql() ([]byte, error) {
	return __062_link_system_preparation_sql, nil
}

var __063_workitem_related_changes_sql = []byte(`-- add a column to record the timestamp of the latest addition/change/removal of an entity in relationship with a workitem
ALTER TABLE work_items ADD COLUMN relationships_changed_at timestamp with time zone;
COMMENT ON COLUMN work_items.relationships_changed_at IS 'see triggers on the ''comments'' and ''work_item_links tables''.';

CREATE FUNCTION workitem_comment_insert_timestamp() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when a comment is added
    BEGIN
        UPDATE work_items wi SET relationships_changed_at = NEW.created_at WHERE wi.id::text = NEW.parent_id;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION workitem_comment_softdelete_timestamp() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`commented_at`+"`"+` column when a comment is removed (soft delete, it's a record update)
    BEGIN
        UPDATE work_items wi SET relationships_changed_at = NEW.deleted_at WHERE wi.id::text = NEW.parent_id;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_comment_insert_trigger AFTER INSERT ON comments 
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE PROCEDURE workitem_comment_insert_timestamp();

CREATE TRIGGER workitem_comment_softdelete_trigger AFTER UPDATE OF deleted_at ON comments 
    FOR EACH ROW
     WHEN (NEW.deleted_at IS NOT NULL)
    EXECUTE PROCEDURE workitem_comment_softdelete_timestamp();
    

CREATE FUNCTION workitem_link_insert_timestamp() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when a link is added
    BEGIN
        UPDATE work_items wi SET relationships_changed_at = NEW.created_at WHERE wi.id in (NEW.source_id, NEW.target_id);
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION workitem_link_softdelete_timestamp() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when a link is removed (soft delete, it's a record update)
    BEGIN
        UPDATE work_items wi SET relationships_changed_at = NEW.deleted_at WHERE wi.id in (NEW.source_id, NEW.target_id);
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_link_insert_trigger AFTER INSERT ON work_item_links 
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE PROCEDURE workitem_link_insert_timestamp();
    
CREATE TRIGGER workitem_link_softdelete_trigger AFTER UPDATE OF deleted_at ON work_item_links 
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NOT NULL)
    EXECUTE PROCEDURE workitem_link_softdelete_timestamp();
`)

func _063_workitem_related_changes_sql() ([]byte, error) {
	return __063_workitem_related_changes_sql, nil
}

var __064_remove_link_combinations_sql = []byte(`DROP TABLE work_item_link_type_combinations;
ALTER TABLE work_item_link_types DROP COLUMN target_type_id;
ALTER TABLE work_item_link_types DROP COLUMN source_type_id;`)

func _064_remove_link_combinations_sql() ([]byte, error) {
	return __064_remove_link_combinations_sql, nil
}

var __065_workitem_id_unique_per_space_sql = []byte(`-- first, we ADD a new COLUMN for the 'natural id' with the same values as the 'id'
ALTER TABLE work_items ADD COLUMN "number" integer;
UPDATE work_items SET number = id;

-- then remove existing CONSTRAINTs and TRIGGERs FROM other tables before changing the 'id' COLUMN
ALTER TABLE work_item_links DROP CONSTRAINT work_item_links_source_id_fkey;
ALTER TABLE work_item_links DROP CONSTRAINT work_item_links_target_id_fkey;
ALTER TABLE work_item_revisions DROP CONSTRAINT work_item_revisions_work_items_fk;
ALTER TABLE work_items DROP CONSTRAINT work_items_pkey;
DROP TRIGGER upd_tsvector ON work_items;
DROP FUNCTION IF EXISTS workitem_tsv_TRIGGER() CASCADE;
DROP INDEX IF EXISTS fulltext_search_index;


-- RENAME COLUMNs of other tables referencing a work item id
ALTER TABLE work_item_links RENAME COLUMN "source_id" TO "source_id_old";
ALTER TABLE work_item_links RENAME COLUMN "target_id" TO "target_id_old";
ALTER TABLE work_item_link_revisions RENAME COLUMN "work_item_link_source_id" TO "work_item_link_source_id_old";
ALTER TABLE work_item_link_revisions RENAME COLUMN "work_item_link_target_id" TO "work_item_link_target_id_old";
ALTER TABLE work_item_revisions RENAME COLUMN "work_item_id" TO "work_item_id_old";

-- ADD new COLUMNs
ALTER TABLE work_item_links ADD COLUMN "source_id" UUID;
ALTER TABLE work_item_links ADD COLUMN "target_id" UUID;
ALTER TABLE work_item_link_revisions ADD COLUMN "work_item_link_source_id" UUID;
ALTER TABLE work_item_link_revisions ADD COLUMN "work_item_link_target_id" UUID;
ALTER TABLE work_item_revisions ADD COLUMN "work_item_id" UUID;

-- assign new UUIDs TO the existing work items
ALTER TABLE work_items DROP COLUMN "id";
ALTER TABLE work_items ADD COLUMN "id" UUID default uuid_generate_v4();
UPDATE work_items SET id = uuid_generate_v4();
-- apply generated UUIDs in other tables
UPDATE work_item_links SET source_id = w.id FROM work_items w WHERE w.number = work_item_links.source_id_old;
UPDATE work_item_links SET target_id = w.id FROM work_items w WHERE w.number = work_item_links.target_id_old;
UPDATE work_item_link_revisions SET work_item_link_source_id = w.id FROM work_items w WHERE w.number = work_item_link_revisions.work_item_link_source_id_old;
UPDATE work_item_link_revisions SET work_item_link_target_id = w.id FROM work_items w WHERE w.number = work_item_link_revisions.work_item_link_target_id_old;
UPDATE work_item_revisions SET work_item_id = w.id FROM work_items w WHERE w.number = work_item_revisions.work_item_id_old;
UPDATE comments SET parent_id = w.id FROM work_items w WHERE w.number::text = comments.parent_id;

-- Drop old columns
ALTER TABLE work_item_links DROP COLUMN "source_id_old";
ALTER TABLE work_item_links DROP COLUMN "target_id_old";
ALTER TABLE work_item_revisions DROP COLUMN "work_item_id_old";
ALTER TABLE work_item_link_revisions DROP COLUMN "work_item_link_source_id_old";
ALTER TABLE work_item_link_revisions DROP COLUMN "work_item_link_target_id_old";

-- recreate constraints, FK and triggers
ALTER TABLE work_items ADD CONSTRAINT work_items_pkey PRIMARY KEY (id);
ALTER TABLE work_item_links ADD CONSTRAINT work_item_links_source_id_fkey FOREIGN KEY (source_id) REFERENCES work_items(id) ON DELETE CASCADE;
ALTER TABLE work_item_links ADD CONSTRAINT work_item_links_target_id_fkey FOREIGN KEY (target_id) REFERENCES work_items(id) ON DELETE CASCADE;
ALTER TABLE work_item_revisions ADD CONSTRAINT work_item_revisions_work_items_fk FOREIGN KEY (work_item_id) REFERENCES work_items(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX work_item_links_unique_idx ON work_item_links (source_id, target_id, link_type_id) WHERE deleted_at IS NULL;

-- create the work item number sequences table
CREATE TABLE work_item_number_sequences (
    space_id uuid primary key,
    current_val integer not null
);
ALTER TABLE work_item_number_sequences ADD CONSTRAINT "work_item_number_sequences_space_id_fkey" FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE;

-- fill the work item ID sequence table with the current 'max' value of issue 'number'
INSERT INTO work_item_number_sequences (space_id, current_val) (select space_id, max(number) FROM work_items group by space_id);

-- ADD unique index ON the work_items table: a 'number' is unique per 'space_id' and those 2 COLUMNs are used TO look-up work items
CREATE UNIQUE INDEX uix_work_items_spaceid_number ON work_items using btree (space_id, number);


-- Restore search capabilities
CREATE INDEX fulltext_search_index ON work_items USING GIN (tsv);

-- UPDATE the 'tsv' COLUMN with the text value of the existing 'content' 
-- element in the 'system.description' JSON document
UPDATE work_items SET tsv =
    setweight(to_tsvector('english', "number"::text),'A') ||
    setweight(to_tsvector('english', coalesce(fields->>'system.title','')),'B') ||
    setweight(to_tsvector('english', coalesce(fields#>>'{system.description, content}','')),'C');

-- fill the 'tsv' COLUMN with the text value of the created/modified 'content' 
-- element in the 'system.description' JSON document
CREATE FUNCTION workitem_tsv_TRIGGER() RETURNS TRIGGER AS $$
begin
  new.tsv :=
    setweight(to_tsvector('english', new.number::text),'A') ||
    setweight(to_tsvector('english', coalesce(new.fields->>'system.title','')),'B') ||
    setweight(to_tsvector('english', coalesce(new.fields#>>'{system.description, content}','')),'C');
  return new;
end
$$ LANGUAGE plpgsql; 

CREATE TRIGGER upd_tsvector BEFORE INSERT OR UPDATE OF number, fields ON work_items
FOR EACH ROW EXECUTE PROCEDURE workitem_tsv_TRIGGER();`)

func _065_workitem_id_unique_per_space_sql() ([]byte, error) {
	return __065_workitem_id_unique_per_space_sql, nil
}

var __066_work_item_links_data_integrity_sql = []byte(`--- remove any record in the 'work_item_links' table if the 'link_type_id', 'source_id' and 'target_id' columns contain `+"`"+`NULL`+"`"+` values
delete from work_item_links where link_type_id IS NULL or source_id IS NULL or target_id IS NULL;

--- make the 'link_type_id', 'source_id' and 'target_id' columns not nullable in the 'work_item_links' table
alter table work_item_links alter column link_type_id set not null;
alter table work_item_links alter column source_id set not null;
alter table work_item_links alter column target_id set not null;`)

func _066_work_item_links_data_integrity_sql() ([]byte, error) {
	return __066_work_item_links_data_integrity_sql, nil
}

var __067_comment_parentid_uuid_sql = []byte(`-- first, we ADD a new COLUMN for the 'parent id' as a UUID in the `+"`"+`comments`+"`"+` table:
ALTER TABLE comments ADD COLUMN "parent_id_uuid" UUID;
UPDATE comments SET parent_id_uuid = parent_id::uuid;
-- then drop the old 'parent_id' column and rename the new one to 'parent_id'
ALTER TABLE comments DROP COLUMN "parent_id";
ALTER TABLE comments RENAME COLUMN "parent_id_uuid" TO "parent_id";

-- second, we ADD a new COLUMN for the 'parent id' as a UUID in the `+"`"+`comment_revisions`+"`"+` table (after migrating the content of `+"`"+`comment_parent_id`+"`"+`, forgotten in step 65) :
ALTER TABLE comment_revisions ADD COLUMN "comment_parent_id_uuid" UUID;
UPDATE comment_revisions SET comment_parent_id_uuid = c.parent_id FROM comments c WHERE c.id = comment_revisions.comment_id;
-- then drop the old 'parent_id' column and rename the new one to 'parent_id'
ALTER TABLE comment_revisions DROP COLUMN "comment_parent_id";
ALTER TABLE comment_revisions RENAME COLUMN "comment_parent_id_uuid" TO "comment_parent_id";

-- also, we need to update the triggers that record the 'relationships_changed_at' value in the 'work_items' table:
DROP TRIGGER workitem_comment_insert_trigger ON comments;
DROP TRIGGER workitem_comment_softdelete_trigger ON comments;
DROP FUNCTION workitem_comment_insert_timestamp();
DROP FUNCTION workitem_comment_softdelete_timestamp();

CREATE FUNCTION workitem_comment_insert_timestamp() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when a comment is added
    BEGIN
        UPDATE work_items wi SET relationships_changed_at = NEW.created_at WHERE wi.id = NEW.parent_id;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION workitem_comment_softdelete_timestamp() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`commented_at`+"`"+` column when a comment is removed (soft delete, it's a record update)
    BEGIN
        UPDATE work_items wi SET relationships_changed_at = NEW.deleted_at WHERE wi.id = NEW.parent_id;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_comment_insert_trigger AFTER INSERT ON comments 
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE PROCEDURE workitem_comment_insert_timestamp();

CREATE TRIGGER workitem_comment_softdelete_trigger AFTER UPDATE OF deleted_at ON comments 
    FOR EACH ROW
     WHEN (NEW.deleted_at IS NOT NULL)
    EXECUTE PROCEDURE workitem_comment_softdelete_timestamp();`)

func _067_comment_parentid_uuid_sql() ([]byte, error) {
	return __067_comment_parentid_uuid_sql, nil
}

var __068_index_identities_username_sql = []byte(`-- add index on identities.username
drop index idx_identities_username;
create index idx_identities_username on identities (lower(username));

`)

func _068_index_identities_username_sql() ([]byte, error) {
	return __068_index_identities_username_sql, nil
}

var __069_limit_execution_order_to_space_sql = []byte(`SELECT 1;
`)

func _069_limit_execution_order_to_space_sql() ([]byte, error) {
	return __069_limit_execution_order_to_space_sql, nil
}

var __070_rename_comment_createdby_to_creator_sql = []byte(`ALTER TABLE comments RENAME created_by TO creator;
`)

func _070_rename_comment_createdby_to_creator_sql() ([]byte, error) {
	return __070_rename_comment_createdby_to_creator_sql, nil
}

var __071_iteration_related_changes_sql = []byte(`ALTER TABLE iterations ADD COLUMN relationships_changed_at timestamp with time zone;
COMMENT ON COLUMN iterations.relationships_changed_at IS 'see triggers on the ''work_items'' table''.';

drop trigger if exists workitem_link_iteration_trigger on work_items;
drop function if exists iteration_set_relationship_timestamp_on_workitem_linking();
drop trigger if exists workitem_unlink_iteration_trigger on work_items;
drop function if exists iteration_set_relationship_timestamp_on_workitem_unlinking();
drop trigger if exists workitem_soft_delete_trigger on work_items;
drop function if exists iteration_set_relationship_timestamp_on_workitem_deletion();

-- trigger and function when a workitem is linked to an iteration
CREATE FUNCTION iteration_set_relationship_timestamp_on_workitem_linking() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when an interation is set
    BEGIN
        UPDATE iterations i SET relationships_changed_at = NEW.updated_at WHERE i.id::text = NEW.Fields->>'system.iteration';
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_link_iteration_trigger AFTER UPDATE ON work_items 
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL 
        -- only occurs when the `+"`"+`system.iteration`+"`"+` field changed to a non-null value
        AND NEW.Fields->>'system.iteration' IS NOT NULL 
        AND (OLD.Fields->>'system.iteration' IS NULL OR NEW.Fields->>'system.iteration' != OLD.Fields->>'system.iteration'))
    EXECUTE PROCEDURE iteration_set_relationship_timestamp_on_workitem_linking();

-- trigger and function when an iteration is unset for a workitem 
CREATE FUNCTION iteration_set_relationship_timestamp_on_workitem_unlinking() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when an interation is set
    BEGIN
        UPDATE iterations i SET relationships_changed_at = NEW.updated_at WHERE i.id::text = OLD.Fields->>'system.iteration';
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_unlink_iteration_trigger AFTER UPDATE ON work_items 
    FOR EACH ROW
    WHEN (OLD.deleted_at IS NULL 
        -- only occurs when the `+"`"+`system.iteration`+"`"+` field was a non-null value before, and then it changed
        AND OLD.Fields->>'system.iteration' IS NOT NULL 
        AND (NEW.Fields->>'system.iteration' IS NULL OR NEW.Fields->>'system.iteration'!= OLD.Fields->>'system.iteration'))
    EXECUTE PROCEDURE iteration_set_relationship_timestamp_on_workitem_unlinking();

-- trigger and function when a workitem that is soft-deleted was linked to an iteration
CREATE FUNCTION iteration_set_relationship_timestamp_on_workitem_deletion() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when an interation is set
    BEGIN
        UPDATE iterations i SET relationships_changed_at = NEW.deleted_at WHERE i.id::text = OLD.Fields->>'system.iteration';
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER workitem_soft_delete_trigger AFTER UPDATE ON work_items 
    FOR EACH ROW
    WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
    EXECUTE PROCEDURE iteration_set_relationship_timestamp_on_workitem_deletion();


`)

func _071_iteration_related_changes_sql() ([]byte, error) {
	return __071_iteration_related_changes_sql, nil
}

var __072_adds_active_flag_in_iteration_sql = []byte(`ALTER TABLE iterations ADD COLUMN user_active bool DEFAULT false NOT NULL;
`)

func _072_adds_active_flag_in_iteration_sql() ([]byte, error) {
	return __072_adds_active_flag_in_iteration_sql, nil
}

var __073_labels_sql = []byte(`CREATE TABLE labels (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    name text NOT NULL CHECK(name <> ''),
    text_color text NOT NULL DEFAULT '#000000' CHECK(text_color ~ '^#[A-Fa-f0-9]{6}$'),
    background_color text NOT NULL DEFAULT '#FFFFFF' CHECK(background_color ~ '^#[A-Fa-f0-9]{6}$'),
    space_id uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    version integer DEFAULT 0 NOT NULL,
    CONSTRAINT labels_name_space_id_unique UNIQUE(space_id, name)
);

CREATE INDEX label_name_idx ON labels USING btree (name);
`)

func _073_labels_sql() ([]byte, error) {
	return __073_labels_sql, nil
}

var __074_label_border_color_sql = []byte(`ALTER TABLE labels ADD COLUMN border_color TEXT NOT NULL DEFAULT '#000000' CHECK(border_color ~ '^#[A-Fa-f0-9]{6}$');
`)

func _074_label_border_color_sql() ([]byte, error) {
	return __074_label_border_color_sql, nil
}

var __075_label_unique_name_sql = []byte(`-- drop existing unique constraint
ALTER table labels DROP CONSTRAINT labels_name_space_id_unique;
-- create unique index on two columns
CREATE UNIQUE INDEX labels_name_space_id_unique_idx ON labels (space_id, name) WHERE deleted_at IS NULL;
`)

func _075_label_unique_name_sql() ([]byte, error) {
	return __075_label_unique_name_sql, nil
}

var __076_drop_space_resources_and_oauth_state_sql = []byte(`-- Space resource management and login have moved to Auth service.
DROP TABLE space_resources;
DROP TABLE oauth_state_references;`)

func _076_drop_space_resources_and_oauth_state_sql() ([]byte, error) {
	return __076_drop_space_resources_and_oauth_state_sql, nil
}

var __077_index_work_item_links_sql = []byte(`-- add an index on work item links to find existing links where a given work item 
-- is the target
CREATE INDEX work_item_links_target_id_idx ON work_item_links USING btree (target_id, deleted_at);`)

func _077_index_work_item_links_sql() ([]byte, error) {
	return __077_index_work_item_links_sql, nil
}

var __078_assignee_and_label_empty_value_sql = []byte(`-- Remove 'system.assignees' from the fields if the value is 'null' or '[]'
update work_items set fields=fields - '{{index . 0}}' where Fields->>'{{index . 0}}' is null;
update work_items set fields=fields - '{{index . 0}}' where Fields->>'{{index . 0}}'='[]';
-- Remove 'system.labels' from the fields if the value is 'null' or '[]'
update work_items set fields=fields - '{{index . 1}}' where Fields->>'{{index . 1}}' is null;
update work_items set fields=fields - '{{index . 1}}' where Fields->>'{{index . 1}}'='[]';
`)

func _078_assignee_and_label_empty_value_sql() ([]byte, error) {
	return __078_assignee_and_label_empty_value_sql, nil
}

var __078_tracker_to_use_uuid_sql = []byte(`-- Add a new UUID field to the trackers table and let other tables use that instead of the current id
ALTER TABLE trackers ADD COLUMN tracker_id uuid DEFAULT uuid_generate_v4() NOT NULL;

-- "Change type" of tracker_id column in tracker_items table
ALTER TABLE tracker_items ADD COLUMN tracker_id_new uuid;
UPDATE tracker_items SET tracker_id_new = trackers.tracker_id FROM trackers WHERE trackers.id = tracker_items.tracker_id;
ALTER TABLE tracker_items ALTER COLUMN tracker_id_new SET NOT NULL;
ALTER TABLE tracker_items DROP COLUMN tracker_id CASCADE;
ALTER TABLE tracker_items RENAME COLUMN tracker_id_new TO tracker_id;

-- "Change type" of tracker_id column in tacker_queries table
ALTER TABLE tracker_queries ADD COLUMN tracker_id_new uuid;
UPDATE tracker_queries SET tracker_id_new = trackers.tracker_id FROM trackers WHERE trackers.id = tracker_queries.tracker_id;
ALTER TABLE tracker_queries ALTER COLUMN tracker_id_new SET NOT NULL;
ALTER TABLE tracker_queries DROP COLUMN tracker_id CASCADE;

-- "Rename" primary key of trackers table
ALTER TABLE tracker_queries RENAME COLUMN tracker_id_new TO tracker_id;
ALTER TABLE trackers DROP COLUMN id CASCADE;
ALTER TABLE trackers RENAME COLUMN tracker_id TO id;
ALTER TABLE trackers ADD CONSTRAINT trackers_pkey PRIMARY KEY (id);

-- Set new foreign keys in tracker_items and tracker_queries to use new UUID field
ALTER TABLE ONLY tracker_items ADD CONSTRAINT tracker_items_tracker_id_trackers_id_foreign FOREIGN KEY (tracker_id) REFERENCES trackers(id) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE ONLY tracker_queries ADD CONSTRAINT tracker_queries_tracker_id_trackers_id_foreign FOREIGN KEY (tracker_id) REFERENCES trackers(id) ON UPDATE RESTRICT ON DELETE RESTRICT;
`)

func _078_tracker_to_use_uuid_sql() ([]byte, error) {
	return __078_tracker_to_use_uuid_sql, nil
}

var __079_assignee_and_label_empty_value_sql = []byte(`-- Remove 'system.assignees' from the fields if the value is 'null' or '[]'
update work_items set fields=fields - '{{index . 0}}' where Fields->>'{{index . 0}}' is null;
update work_items set fields=fields - '{{index . 0}}' where Fields->>'{{index . 0}}'='[]';
-- Remove 'system.labels' from the fields if the value is 'null' or '[]'
update work_items set fields=fields - '{{index . 1}}' where Fields->>'{{index . 1}}' is null;
update work_items set fields=fields - '{{index . 1}}' where Fields->>'{{index . 1}}'='[]';
`)

func _079_assignee_and_label_empty_value_sql() ([]byte, error) {
	return __079_assignee_and_label_empty_value_sql, nil
}

var __080_remove_unknown_link_types_sql = []byte(`SET LOCAL linktypes.bug_blocker = '{{index . 0}}';
SET LOCAL linktypes.related = '{{index . 1}}';
SET LOCAL linktypes.parenting = '{{index . 2}}';

SET LOCAL linkcats.systemcat = '{{index . 3}}';
SET LOCAL linkcats.usercat = '{{index . 4}}';

-- The following link types exist in the current production database but they
-- are not known to the code and therefore we re-assign all links associated to
-- those link types with their appropriate link type and later remove these
-- unknown link types.
SET LOCAL linktypes.unknown_bug_blocker = 'aad2a4ad-d601-4104-9804-2c977ca2e0c1';
SET LOCAL linktypes.unknown_related = '355b647b-adc5-46b3-b297-cc54bc0554e6';
SET LOCAL linktypes.unknown_parenting = '7479a9b9-8607-46fa-9535-d448fa8768ab';

-- Re-assign any existing link to correct type but avoid those links that
-- already exist.
UPDATE work_item_links l 
    SET link_type_id = current_setting('linktypes.bug_blocker')::uuid 
    WHERE link_type_id = current_setting('linktypes.unknown_bug_blocker')::uuid
    AND NOT EXISTS(
        SELECT * FROM work_item_links
        WHERE
            link_type_id = current_setting('linktypes.bug_blocker')::uuid
            AND source_id = l.source_id
            AND target_id = l.target_id
        );
UPDATE work_item_links l 
    SET link_type_id = current_setting('linktypes.related')::uuid 
    WHERE link_type_id = current_setting('linktypes.unknown_related')::uuid
    AND NOT EXISTS(
        SELECT * FROM work_item_links
        WHERE
            link_type_id = current_setting('linktypes.related')::uuid
            AND source_id = l.source_id
            AND target_id = l.target_id
        );
UPDATE work_item_links l 
    SET link_type_id = current_setting('linktypes.parenting')::uuid 
    WHERE link_type_id = current_setting('linktypes.unknown_parenting')::uuid
    AND NOT EXISTS(
        SELECT * FROM work_item_links
        WHERE
            link_type_id = current_setting('linktypes.parenting')::uuid
            AND source_id = l.source_id
            AND target_id = l.target_id
        );

-- Update revisions
UPDATE work_item_link_revisions rev SET
    work_item_link_type_id = (SELECT link_type_id FROM work_item_links WHERE id = rev.work_item_link_id);

-- Remove unknown link categories
DELETE FROM work_item_link_categories WHERE id NOT IN (
     current_setting('linkcats.systemcat')::uuid,
     current_setting('linkcats.usercat')::uuid
);

-- Finally, delete old link types
DELETE FROM work_item_link_types WHERE id NOT IN (
    current_setting('linktypes.bug_blocker')::uuid,
    current_setting('linktypes.related')::uuid,
    current_setting('linktypes.parenting')::uuid
);

-- Add foreign keys to revisions table
ALTER TABLE work_item_link_revisions  ADD CONSTRAINT link_rev_link_type_fk  FOREIGN KEY (work_item_link_type_id) REFERENCES work_item_link_types(id) ON DELETE CASCADE;
ALTER TABLE work_item_link_revisions ADD CONSTRAINT link_rev_source_id_fk FOREIGN KEY (work_item_link_source_id) REFERENCES work_items(id) ON DELETE CASCADE;
ALTER TABLE work_item_link_revisions ADD CONSTRAINT link_rev_target_id_fk FOREIGN KEY (work_item_link_target_id) REFERENCES work_items(id) ON DELETE CASCADE;`)

func _080_remove_unknown_link_types_sql() ([]byte, error) {
	return __080_remove_unknown_link_types_sql, nil
}

var __081_queries_sql = []byte(`CREATE TABLE queries (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    title text NOT NULL CHECK(title <> ''),
    fields jsonb NOT NULL,
    space_id uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    creator uuid NOT NULL
);
CREATE UNIQUE INDEX queries_title_space_id_creator_unique ON queries (title, space_id, creator) WHERE deleted_at IS NULL;

CREATE INDEX query_creator_idx ON queries USING btree (creator);
`)

func _081_queries_sql() ([]byte, error) {
	return __081_queries_sql, nil
}

var __082_iteration_related_changes_sql = []byte(`-- looks very similar to step 071 but here we replace the 
-- `+"`"+`WHERE i.id::text = NEW.Fields->>'system.iteration'`+"`"+` comparison with 
-- `+"`"+`WHERE i.id = (NEW.Fields->>'system.iteration')::uuid`+"`"+` to use the 
-- index of iterations on the `+"`"+`id`+"`"+` column (primary key)
drop trigger if exists workitem_link_iteration_trigger on work_items;
drop function if exists iteration_set_relationship_timestamp_on_workitem_linking();
drop trigger if exists workitem_unlink_iteration_trigger on work_items;
drop function if exists iteration_set_relationship_timestamp_on_workitem_unlinking();
drop trigger if exists workitem_soft_delete_trigger on work_items;
drop function if exists iteration_set_relationship_timestamp_on_workitem_deletion();

-- trigger and function when a workitem is linked to an iteration
CREATE FUNCTION iteration_set_relationship_timestamp_on_workitem_linking() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when an interation is set
    BEGIN
        UPDATE iterations i SET relationships_changed_at = NEW.updated_at 
        WHERE NEW.Fields->>'system.iteration' IS NOT NULL AND i.id = (NEW.Fields->>'system.iteration')::uuid;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_link_iteration_trigger AFTER UPDATE ON work_items 
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL 
        -- only occurs when the `+"`"+`system.iteration`+"`"+` field changed to a non-null value
        AND NEW.Fields->>'system.iteration' IS NOT NULL 
        AND (OLD.Fields->>'system.iteration' IS NULL OR NEW.Fields->>'system.iteration' != OLD.Fields->>'system.iteration'))
    EXECUTE PROCEDURE iteration_set_relationship_timestamp_on_workitem_linking();

-- trigger and function when an iteration is unset for a workitem 
CREATE FUNCTION iteration_set_relationship_timestamp_on_workitem_unlinking() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when an interation is set
    BEGIN
        UPDATE iterations i SET relationships_changed_at = NEW.updated_at 
        WHERE OLD.Fields->>'system.iteration' IS NOT NULL AND i.id = (OLD.Fields->>'system.iteration')::uuid;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER workitem_unlink_iteration_trigger AFTER UPDATE ON work_items 
    FOR EACH ROW
    WHEN (OLD.deleted_at IS NULL 
        -- only occurs when the `+"`"+`system.iteration`+"`"+` field was a non-null value before, and then it changed
        AND OLD.Fields->>'system.iteration' IS NOT NULL 
        AND (NEW.Fields->>'system.iteration' IS NULL OR NEW.Fields->>'system.iteration'!= OLD.Fields->>'system.iteration'))
    EXECUTE PROCEDURE iteration_set_relationship_timestamp_on_workitem_unlinking();

-- trigger and function when a workitem that is soft-deleted was linked to an iteration
CREATE FUNCTION iteration_set_relationship_timestamp_on_workitem_deletion() RETURNS trigger AS $$
    -- trigger to fill the `+"`"+`relationships_changed_at`+"`"+` column when an interation is set
    BEGIN
        UPDATE iterations i SET relationships_changed_at = NEW.deleted_at 
        WHERE OLD.Fields->>'system.iteration' IS NOT NULL AND i.id = (OLD.Fields->>'system.iteration')::uuid;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER workitem_soft_delete_trigger AFTER UPDATE ON work_items 
    FOR EACH ROW
    WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
    EXECUTE PROCEDURE iteration_set_relationship_timestamp_on_workitem_deletion();


`)

func _082_iteration_related_changes_sql() ([]byte, error) {
	return __082_iteration_related_changes_sql, nil
}

var __083_index_comments_parent_sql = []byte(`-- add index on comments.parent_id
create index idx_comments_parentid on comments using btree (parent_id);

`)

func _083_index_comments_parent_sql() ([]byte, error) {
	return __083_index_comments_parent_sql, nil
}

var __084_codebases_spaceid_url_index_sql = []byte(`-- Delete duplicate entries of codebases in same space
-- See here: https://wiki.postgresql.org/wiki/Deleting_duplicates
DELETE FROM codebases
WHERE id IN (
    SELECT id
    FROM (
        SELECT id, ROW_NUMBER() OVER (partition BY url, space_id ORDER BY deleted_at DESC) AS rnum
        FROM codebases
    ) t
    WHERE t.rnum > 1
);

-- From now on ensure we have codebase only once in space
CREATE UNIQUE INDEX codebases_spaceid_url_idx
ON codebases (url, space_id) WHERE deleted_at IS NULL;
`)

func _084_codebases_spaceid_url_index_sql() ([]byte, error) {
	return __084_codebases_spaceid_url_index_sql, nil
}

var __085_delete_system_number_json_field_sql = []byte(`-- This removes any potentially existing system.number field from all work items.
UPDATE work_items SET fields=(fields - 'system.number');
`)

func _085_delete_system_number_json_field_sql() ([]byte, error) {
	return __085_delete_system_number_json_field_sql, nil
}

var __086_add_can_construct_to_wit_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

SET LOCAL idx.planner_item_type_id = '{{index . 0}}';

-- Add boolean can_construct field to work item type table and make it default
-- to TRUE.
ALTER TABLE work_item_types ADD COLUMN can_construct boolean;
UPDATE work_item_types SET can_construct = TRUE;
UPDATE work_item_types SET can_construct = FALSE WHERE id = current_setting('idx.planner_item_type_id')::uuid;
ALTER TABLE work_item_types ALTER can_construct SET DEFAULT TRUE;
ALTER TABLE work_item_types ALTER COLUMN can_construct SET NOT NULL;`)

func _086_add_can_construct_to_wit_sql() ([]byte, error) {
	return __086_add_can_construct_to_wit_sql, nil
}

var __087_space_templates_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

SET LOCAL idx.space_template_id = '{{index . 0}}';
SET LOCAL idx.planner_item_type_id = '{{index . 1}}';

-- Remove space_id field from link types (WILTs) and work item types (WITs) This
-- can be done because all WILTs and WITs exist in the system space anyway. So
-- in order to maintain a compatibility with the current API on controller level
-- we can just fake the space relationship to be pointing to the system space.
ALTER TABLE work_item_link_types DROP COLUMN space_id;
ALTER TABLE work_item_types DROP COLUMN space_id;

CREATE TABLE space_templates (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    version integer DEFAULT 0 NOT NULL,
    name text NOT NULL CHECK(name <> ''),
    description text,
    can_construct boolean DEFAULT TRUE NOT NULL
);
CREATE UNIQUE INDEX space_templates_name_uidx ON space_templates (name) WHERE deleted_at IS NULL;

-- Create a default empty space template
INSERT INTO space_templates (id, name, description) VALUES(
    current_setting('idx.space_template_id')::uuid,
    'empty space template',
    'this will be overwritten by the legacy space template when common types are populated'
);

-- Add foreign key to spaces relation
ALTER TABLE spaces ADD COLUMN space_template_id uuid REFERENCES space_templates(id) ON DELETE CASCADE;
UPDATE spaces SET space_template_id = current_setting('idx.space_template_id')::uuid;
ALTER TABLE spaces ALTER COLUMN space_template_id SET NOT NULL;

-- Add foreign key to work item type relation
ALTER TABLE work_item_types ADD COLUMN space_template_id uuid REFERENCES space_templates(id) ON DELETE CASCADE;
UPDATE work_item_types SET space_template_id = current_setting('idx.space_template_id')::uuid;
ALTER TABLE work_item_types ALTER COLUMN space_template_id SET NOT NULL;

-- Add foreign key to work item link type relation
ALTER TABLE work_item_link_types ADD COLUMN space_template_id uuid REFERENCES space_templates(id) ON DELETE CASCADE;
UPDATE work_item_link_types SET space_template_id = current_setting('idx.space_template_id')::uuid;
ALTER TABLE work_item_link_types ALTER COLUMN space_template_id SET NOT NULL;
`)

func _087_space_templates_sql() ([]byte, error) {
	return __087_space_templates_sql, nil
}

var __088_type_groups_and_child_types_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE type_group_bucket_enum AS ENUM('portfolio', 'requirement', 'iteration');

CREATE TABLE work_item_type_groups (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    version integer DEFAULT 0 NOT NULL,
    position integer DEFAULT 0 NOT NULL,
    name text NOT NULL CHECK(name <> ''),
    bucket type_group_bucket_enum NOT NULL,
    icon text,
    space_template_id uuid REFERENCES space_templates(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX work_item_type_groups_name_uidx ON work_item_type_groups (name, space_template_id) WHERE deleted_at IS NULL;

CREATE TABLE work_item_type_group_members (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_group_id uuid REFERENCES work_item_type_groups(id) ON DELETE CASCADE,
    work_item_type_id uuid REFERENCES work_item_types(id) ON DELETE CASCADE,
    position integer DEFAULT 0 NOT NULL
);

CREATE UNIQUE INDEX work_item_type_group_members_uidx ON work_item_type_group_members (type_group_id, work_item_type_id) WHERE deleted_at IS NULL;

CREATE TABLE work_item_child_types (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_work_item_type_id uuid REFERENCES work_item_types(id) ON DELETE CASCADE,
    child_work_item_type_id uuid REFERENCES work_item_types(id) ON DELETE CASCADE,
    position integer DEFAULT 0 NOT NULL
);

CREATE UNIQUE INDEX work_item_child_types_uidx ON work_item_child_types (parent_work_item_type_id, child_work_item_type_id) WHERE deleted_at IS NULL;

-- Only allow one work item link type with the same name for the same space
-- template in existence.
CREATE UNIQUE INDEX work_item_link_types_name_idx ON work_item_link_types (name, space_template_id) WHERE deleted_at IS NULL;`)

func _088_type_groups_and_child_types_sql() ([]byte, error) {
	return __088_type_groups_and_child_types_sql, nil
}

var __089_fixup_space_templates_sql = []byte(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

SET LOCAL idx.legacy_space_template_id = '{{index . 0}}';
SET LOCAL idx.base_space_template_id = '{{index . 1}}';
SET LOCAL idx.planner_item_type_id = '{{index . 2}}';

-- create base space template
INSERT INTO space_templates (id, name, description) VALUES(
    current_setting('idx.base_space_template_id')::uuid,
    'base space template',
    'this will be overwritten by the base space template when common types are populated'
);

-- Add foreign key to work item type relation and make all but the planner item
-- type a part of the base template.
UPDATE work_item_types SET space_template_id = current_setting('idx.base_space_template_id')::uuid WHERE id = current_setting('idx.planner_item_type_id')::uuid;

UPDATE work_item_types SET space_template_id = current_setting('idx.legacy_space_template_id')::uuid WHERE id <> current_setting('idx.planner_item_type_id')::uuid;

-- Add foreign key to work item link type relation and make all existing link
-- types a part of the base template.
UPDATE work_item_link_types SET space_template_id = current_setting('idx.base_space_template_id')::uuid;
`)

func _089_fixup_space_templates_sql() ([]byte, error) {
	return __089_fixup_space_templates_sql, nil
}

var __090_queries_version_sql = []byte(`ALTER TABLE queries ADD COLUMN version INTEGER DEFAULT 0 NOT NULL;
`)

func _090_queries_version_sql() ([]byte, error) {
	return __090_queries_version_sql, nil
}

var __091_comments_child_comments_sql = []byte(`ALTER TABLE comments ADD COLUMN parent_comment_id uuid;
ALTER TABLE comments ADD CONSTRAINT comments_parent_comment_id_comment_id_fk FOREIGN KEY (parent_comment_id) REFERENCES comments (id);
`)

func _091_comments_child_comments_sql() ([]byte, error) {
	return __091_comments_child_comments_sql, nil
}

var __092_comment_revisions_child_comments_sql = []byte(`ALTER TABLE comment_revisions ADD COLUMN comment_parent_comment_id uuid;`)

func _092_comment_revisions_child_comments_sql() ([]byte, error) {
	return __092_comment_revisions_child_comments_sql, nil
}

var __093_codebase_add_cve_scan_sql = []byte(`-- Add new column to codebases cve_scan which keeps track
-- whether this particular codebase should be scanned or not
ALTER TABLE codebases ADD COLUMN cve_scan BOOLEAN;
UPDATE codebases SET cve_scan = 'f';
ALTER TABLE codebases ALTER COLUMN cve_scan SET NOT NULL;
ALTER TABLE codebases ALTER COLUMN cve_scan SET DEFAULT TRUE;
`)

func _093_codebase_add_cve_scan_sql() ([]byte, error) {
	return __093_codebase_add_cve_scan_sql, nil
}

var __094_changes_to_agile_template_sql = []byte(`
-- This removes any potentially existing effort field from all work items
-- and the work item type definition of the type theme.
UPDATE work_items SET fields=(fields - 'effort') WHERE type='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
UPDATE work_item_types SET fields=(fields - 'effort') WHERE id='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';

-- This removes any potentially existing business_value field from all work items
-- and the work item type definition of the type theme.
UPDATE work_items SET fields=(fields - 'business_value') WHERE type='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
UPDATE work_item_types SET fields=(fields - 'business_value') WHERE id='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';

-- This removes any potentially existing time_criticality field from all work items
-- and the work item type definition of the type theme.
UPDATE work_items SET fields=(fields - 'time_criticality') WHERE type='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
UPDATE work_item_types SET fields=(fields - 'time_criticality') WHERE id='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';

-- This removes any potentially existing effort field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'effort') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'effort') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing business_value field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'business_value') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'business_value') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing time_criticality field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'time_criticality') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'time_criticality') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing component field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'component') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'component') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing effort field from all work items
-- and the work item type definition of the type story.
UPDATE work_items SET fields=(fields - 'effort') WHERE type='6ff83406-caa7-47a9-9200-4ca796be11bb';
UPDATE work_item_types SET fields=(fields - 'effort') WHERE id='6ff83406-caa7-47a9-9200-4ca796be11bb';
`)

func _094_changes_to_agile_template_sql() ([]byte, error) {
	return __094_changes_to_agile_template_sql, nil
}

var __095_remove_resolution_field_from_impediment_sql = []byte(`-- This removes any potentially existing "resolution" field from all
-- "impediment" work items and the work item type definition of the
-- "impediment". This is needed because a recent space template change
-- (https://github.com/fabric8-services/fabric8-wit/pull/2133) removed or
-- switched the order of the values in this impediment type and therefore
-- couldn't be applied. When we remove the "resolution" field here it here and
-- then import the agile space template, it will create a new "resolution" field
-- on the "impediment" work item type.
--
-- (for an error description see
-- https://github.com/openshiftio/openshift.io/issues/3879)
UPDATE work_items SET fields=(fields - 'resolution') WHERE type='03b9bb64-4f65-4fa7-b165-494cd4f01401';
UPDATE work_item_types SET fields=(fields - 'resolution') WHERE id='03b9bb64-4f65-4fa7-b165-494cd4f01401';`)

func _095_remove_resolution_field_from_impediment_sql() ([]byte, error) {
	return __095_remove_resolution_field_from_impediment_sql, nil
}

var __096_changes_to_agile_template_sql = []byte(`
-- This removes any potentially existing effort field from all work items
-- and the work item type definition of the type theme.
UPDATE work_items SET fields=(fields - 'effort') WHERE type='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
UPDATE work_item_types SET fields=(fields - 'effort') WHERE id='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';

-- This removes any potentially existing business_value field from all work items
-- and the work item type definition of the type theme.
UPDATE work_items SET fields=(fields - 'business_value') WHERE type='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
UPDATE work_item_types SET fields=(fields - 'business_value') WHERE id='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';

-- This removes any potentially existing time_criticality field from all work items
-- and the work item type definition of the type theme.
UPDATE work_items SET fields=(fields - 'time_criticality') WHERE type='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
UPDATE work_item_types SET fields=(fields - 'time_criticality') WHERE id='5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';

-- This removes any potentially existing effort field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'effort') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'effort') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing business_value field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'business_value') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'business_value') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing time_criticality field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'time_criticality') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'time_criticality') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing component field from all work items
-- and the work item type definition of the type epic.
UPDATE work_items SET fields=(fields - 'component') WHERE type='2c169431-a55d-49eb-af74-cc19e895356f';
UPDATE work_item_types SET fields=(fields - 'component') WHERE id='2c169431-a55d-49eb-af74-cc19e895356f';

-- This removes any potentially existing effort field from all work items
-- and the work item type definition of the type story.
UPDATE work_items SET fields=(fields - 'effort') WHERE type='6ff83406-caa7-47a9-9200-4ca796be11bb';
UPDATE work_item_types SET fields=(fields - 'effort') WHERE id='6ff83406-caa7-47a9-9200-4ca796be11bb';
`)

func _096_changes_to_agile_template_sql() ([]byte, error) {
	return __096_changes_to_agile_template_sql, nil
}

var __097_remove_resolution_field_from_impediment_sql = []byte(`-- This removes any potentially existing "resolution" field from all
-- "impediment" work items and the work item type definition of the
-- "impediment". This is needed because a recent space template change
-- (https://github.com/fabric8-services/fabric8-wit/pull/2133) removed or
-- switched the order of the values in this impediment type and therefore
-- couldn't be applied. When we remove the "resolution" field here it here and
-- then import the agile space template, it will create a new "resolution" field
-- on the "impediment" work item type.
--
-- (for an error description see
-- https://github.com/openshiftio/openshift.io/issues/3879)
UPDATE work_items SET fields=(fields - 'resolution') WHERE type='03b9bb64-4f65-4fa7-b165-494cd4f01401';
UPDATE work_item_types SET fields=(fields - 'resolution') WHERE id='03b9bb64-4f65-4fa7-b165-494cd4f01401';`)

func _097_remove_resolution_field_from_impediment_sql() ([]byte, error) {
	return __097_remove_resolution_field_from_impediment_sql, nil
}

var _readme_adoc = []byte(`= SQL files
:toc:
:toc-placement: preamble
:sectnums:
:experimental:

== Purpose

The purpose of this directory is to store all files that are relevant for
updating the database over time.

The SQL files themselves are packaged into the wit core binary with
link:https://github.com/jteeuwen/go-bindata[go-bindata].

The filenames of the SQL files have no meaning but we prefix them with the
version they stand for so it is easier to find out what's happening.
The link:../migration.go[migration.go] file has the control over the
updates and the SQL files are *not* blindly executed just because they exist.
Instead we allow the developers to run Go code as well.`)

func readme_adoc() ([]byte, error) {
	return _readme_adoc, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"000-bootstrap.sql": _000_bootstrap_sql,
	"001-common.sql": _001_common_sql,
	"002-tracker-items.sql": _002_tracker_items_sql,
	"003-login.sql": _003_login_sql,
	"004-drop-tracker-query-id.sql": _004_drop_tracker_query_id_sql,
	"005-add-search-index.sql": _005_add_search_index_sql,
	"006-rename-parent-path.sql": _006_rename_parent_path_sql,
	"007-work-item-links.sql": _007_work_item_links_sql,
	"008-soft-delete-or-resurrect.sql": _008_soft_delete_or_resurrect_sql,
	"009-drop-wit-trigger.sql": _009_drop_wit_trigger_sql,
	"010-comments.sql": _010_comments_sql,
	"011-projects.sql": _011_projects_sql,
	"012-unique-work-item-links.sql": _012_unique_work_item_links_sql,
	"013-iterations.sql": _013_iterations_sql,
	"014-wi-fields-index.sql": _014_wi_fields_index_sql,
	"015-rename-projects-to-spaces.sql": _015_rename_projects_to_spaces_sql,
	"016-drop-wi-links-trigger.sql": _016_drop_wi_links_trigger_sql,
	"017-alter-iterations.sql": _017_alter_iterations_sql,
	"018-rewrite-wits.sql": _018_rewrite_wits_sql,
	"019-add-state-iterations.sql": _019_add_state_iterations_sql,
	"020-work-item-description-update-search-index.sql": _020_work_item_description_update_search_index_sql,
	"021-add-space-description.sql": _021_add_space_description_sql,
	"022-work-item-description-update.sql": _022_work_item_description_update_sql,
	"023-comment-markup.sql": _023_comment_markup_sql,
	"024-comment-markup-default.sql": _024_comment_markup_default_sql,
	"025-refactor-identities-users.sql": _025_refactor_identities_users_sql,
	"026-areas.sql": _026_areas_sql,
	"027-areas-index.sql": _027_areas_index_sql,
	"028-identity-provider_url.sql": _028_identity_provider_url_sql,
	"029-identities-foreign-key.sql": _029_identities_foreign_key_sql,
	"030-indentities-unique-index.sql": _030_indentities_unique_index_sql,
	"031-iterations-parent-path-ltree.sql": _031_iterations_parent_path_ltree_sql,
	"032-add-foreign-key-space-id.sql": _032_add_foreign_key_space_id_sql,
	"033-add-space-id-wilt.sql": _033_add_space_id_wilt_sql,
	"034-space-owner.sql": _034_space_owner_sql,
	"035-wit-to-use-uuid.sql": _035_wit_to_use_uuid_sql,
	"036-add-icon-to-wit.sql": _036_add_icon_to_wit_sql,
	"037-work-item-revisions.sql": _037_work_item_revisions_sql,
	"038-comment-revisions.sql": _038_comment_revisions_sql,
	"039-comment-revisions-parentid.sql": _039_comment_revisions_parentid_sql,
	"040-add-space-id-wi-wit-tq.sql": _040_add_space_id_wi_wit_tq_sql,
	"041-unique-area-name-create-new-area.sql": _041_unique_area_name_create_new_area_sql,
	"042-work-item-link-revisions.sql": _042_work_item_link_revisions_sql,
	"043-space-resources.sql": _043_space_resources_sql,
	"044-add-contextinfo-column-users.sql": _044_add_contextinfo_column_users_sql,
	"045-adds-order-to-existing-wi.sql": _045_adds_order_to_existing_wi_sql,
	"046-oauth-states.sql": _046_oauth_states_sql,
	"047-codebases.sql": _047_codebases_sql,
	"048-unique-iteration-name-create-new-iteration.sql": _048_unique_iteration_name_create_new_iteration_sql,
	"049-add-wi-to-root-area.sql": _049_add_wi_to_root_area_sql,
	"050-add-company-to-user-profile.sql": _050_add_company_to_user_profile_sql,
	"051-modify-work_item_link_types_name_idx.sql": _051_modify_work_item_link_types_name_idx_sql,
	"052-unique-space-names.sql": _052_unique_space_names_sql,
	"053-edit-username.sql": _053_edit_username_sql,
	"054-add-stackid-to-codebase.sql": _054_add_stackid_to_codebase_sql,
	"055-assign-root-area-if-missing.sql": _055_assign_root_area_if_missing_sql,
	"056-assign-root-iteration-if-missing.sql": _056_assign_root_iteration_if_missing_sql,
	"057-add-last-used-workspace-to-codebase.sql": _057_add_last_used_workspace_to_codebase_sql,
	"058-index-identities-fullname.sql": _058_index_identities_fullname_sql,
	"059-fixed-ids-for-system-link-types-and-categories.sql": _059_fixed_ids_for_system_link_types_and_categories_sql,
	"060-fixed-identities-username-idx.sql": _060_fixed_identities_username_idx_sql,
	"061-replace-index-space-name.sql": _061_replace_index_space_name_sql,
	"062-link-system-preparation.sql": _062_link_system_preparation_sql,
	"063-workitem-related-changes.sql": _063_workitem_related_changes_sql,
	"064-remove-link-combinations.sql": _064_remove_link_combinations_sql,
	"065-workitem-id-unique-per-space.sql": _065_workitem_id_unique_per_space_sql,
	"066-work_item_links_data_integrity.sql": _066_work_item_links_data_integrity_sql,
	"067-comment-parentid-uuid.sql": _067_comment_parentid_uuid_sql,
	"068-index_identities_username.sql": _068_index_identities_username_sql,
	"069-limit-execution-order-to-space.sql": _069_limit_execution_order_to_space_sql,
	"070-rename-comment-createdby-to-creator.sql": _070_rename_comment_createdby_to_creator_sql,
	"071-iteration-related-changes.sql": _071_iteration_related_changes_sql,
	"072-adds-active-flag-in-iteration.sql": _072_adds_active_flag_in_iteration_sql,
	"073-labels.sql": _073_labels_sql,
	"074-label-border-color.sql": _074_label_border_color_sql,
	"075-label-unique-name.sql": _075_label_unique_name_sql,
	"076-drop-space-resources-and-oauth-state.sql": _076_drop_space_resources_and_oauth_state_sql,
	"077-index-work-item-links.sql": _077_index_work_item_links_sql,
	"078-assignee-and-label-empty-value.sql": _078_assignee_and_label_empty_value_sql,
	"078-tracker-to-use-uuid.sql": _078_tracker_to_use_uuid_sql,
	"079-assignee-and-label-empty-value.sql": _079_assignee_and_label_empty_value_sql,
	"080-remove-unknown-link-types.sql": _080_remove_unknown_link_types_sql,
	"081-queries.sql": _081_queries_sql,
	"082-iteration-related-changes.sql": _082_iteration_related_changes_sql,
	"083-index-comments-parent.sql": _083_index_comments_parent_sql,
	"084-codebases-spaceid-url-index.sql": _084_codebases_spaceid_url_index_sql,
	"085-delete-system.number-json-field.sql": _085_delete_system_number_json_field_sql,
	"086-add-can-construct-to-wit.sql": _086_add_can_construct_to_wit_sql,
	"087-space-templates.sql": _087_space_templates_sql,
	"088-type-groups-and-child-types.sql": _088_type_groups_and_child_types_sql,
	"089-fixup-space-templates.sql": _089_fixup_space_templates_sql,
	"090-queries-version.sql": _090_queries_version_sql,
	"091-comments-child-comments.sql": _091_comments_child_comments_sql,
	"092-comment-revisions-child-comments.sql": _092_comment_revisions_child_comments_sql,
	"093-codebase-add-cve-scan.sql": _093_codebase_add_cve_scan_sql,
	"094-changes-to-agile-template.sql": _094_changes_to_agile_template_sql,
	"095-remove-resolution-field-from-impediment.sql": _095_remove_resolution_field_from_impediment_sql,
	"096-changes-to-agile-template.sql": _096_changes_to_agile_template_sql,
	"097-remove-resolution-field-from-impediment.sql": _097_remove_resolution_field_from_impediment_sql,
	"README.adoc": readme_adoc,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"000-bootstrap.sql": &_bintree_t{_000_bootstrap_sql, map[string]*_bintree_t{
	}},
	"001-common.sql": &_bintree_t{_001_common_sql, map[string]*_bintree_t{
	}},
	"002-tracker-items.sql": &_bintree_t{_002_tracker_items_sql, map[string]*_bintree_t{
	}},
	"003-login.sql": &_bintree_t{_003_login_sql, map[string]*_bintree_t{
	}},
	"004-drop-tracker-query-id.sql": &_bintree_t{_004_drop_tracker_query_id_sql, map[string]*_bintree_t{
	}},
	"005-add-search-index.sql": &_bintree_t{_005_add_search_index_sql, map[string]*_bintree_t{
	}},
	"006-rename-parent-path.sql": &_bintree_t{_006_rename_parent_path_sql, map[string]*_bintree_t{
	}},
	"007-work-item-links.sql": &_bintree_t{_007_work_item_links_sql, map[string]*_bintree_t{
	}},
	"008-soft-delete-or-resurrect.sql": &_bintree_t{_008_soft_delete_or_resurrect_sql, map[string]*_bintree_t{
	}},
	"009-drop-wit-trigger.sql": &_bintree_t{_009_drop_wit_trigger_sql, map[string]*_bintree_t{
	}},
	"010-comments.sql": &_bintree_t{_010_comments_sql, map[string]*_bintree_t{
	}},
	"011-projects.sql": &_bintree_t{_011_projects_sql, map[string]*_bintree_t{
	}},
	"012-unique-work-item-links.sql": &_bintree_t{_012_unique_work_item_links_sql, map[string]*_bintree_t{
	}},
	"013-iterations.sql": &_bintree_t{_013_iterations_sql, map[string]*_bintree_t{
	}},
	"014-wi-fields-index.sql": &_bintree_t{_014_wi_fields_index_sql, map[string]*_bintree_t{
	}},
	"015-rename-projects-to-spaces.sql": &_bintree_t{_015_rename_projects_to_spaces_sql, map[string]*_bintree_t{
	}},
	"016-drop-wi-links-trigger.sql": &_bintree_t{_016_drop_wi_links_trigger_sql, map[string]*_bintree_t{
	}},
	"017-alter-iterations.sql": &_bintree_t{_017_alter_iterations_sql, map[string]*_bintree_t{
	}},
	"018-rewrite-wits.sql": &_bintree_t{_018_rewrite_wits_sql, map[string]*_bintree_t{
	}},
	"019-add-state-iterations.sql": &_bintree_t{_019_add_state_iterations_sql, map[string]*_bintree_t{
	}},
	"020-work-item-description-update-search-index.sql": &_bintree_t{_020_work_item_description_update_search_index_sql, map[string]*_bintree_t{
	}},
	"021-add-space-description.sql": &_bintree_t{_021_add_space_description_sql, map[string]*_bintree_t{
	}},
	"022-work-item-description-update.sql": &_bintree_t{_022_work_item_description_update_sql, map[string]*_bintree_t{
	}},
	"023-comment-markup.sql": &_bintree_t{_023_comment_markup_sql, map[string]*_bintree_t{
	}},
	"024-comment-markup-default.sql": &_bintree_t{_024_comment_markup_default_sql, map[string]*_bintree_t{
	}},
	"025-refactor-identities-users.sql": &_bintree_t{_025_refactor_identities_users_sql, map[string]*_bintree_t{
	}},
	"026-areas.sql": &_bintree_t{_026_areas_sql, map[string]*_bintree_t{
	}},
	"027-areas-index.sql": &_bintree_t{_027_areas_index_sql, map[string]*_bintree_t{
	}},
	"028-identity-provider_url.sql": &_bintree_t{_028_identity_provider_url_sql, map[string]*_bintree_t{
	}},
	"029-identities-foreign-key.sql": &_bintree_t{_029_identities_foreign_key_sql, map[string]*_bintree_t{
	}},
	"030-indentities-unique-index.sql": &_bintree_t{_030_indentities_unique_index_sql, map[string]*_bintree_t{
	}},
	"031-iterations-parent-path-ltree.sql": &_bintree_t{_031_iterations_parent_path_ltree_sql, map[string]*_bintree_t{
	}},
	"032-add-foreign-key-space-id.sql": &_bintree_t{_032_add_foreign_key_space_id_sql, map[string]*_bintree_t{
	}},
	"033-add-space-id-wilt.sql": &_bintree_t{_033_add_space_id_wilt_sql, map[string]*_bintree_t{
	}},
	"034-space-owner.sql": &_bintree_t{_034_space_owner_sql, map[string]*_bintree_t{
	}},
	"035-wit-to-use-uuid.sql": &_bintree_t{_035_wit_to_use_uuid_sql, map[string]*_bintree_t{
	}},
	"036-add-icon-to-wit.sql": &_bintree_t{_036_add_icon_to_wit_sql, map[string]*_bintree_t{
	}},
	"037-work-item-revisions.sql": &_bintree_t{_037_work_item_revisions_sql, map[string]*_bintree_t{
	}},
	"038-comment-revisions.sql": &_bintree_t{_038_comment_revisions_sql, map[string]*_bintree_t{
	}},
	"039-comment-revisions-parentid.sql": &_bintree_t{_039_comment_revisions_parentid_sql, map[string]*_bintree_t{
	}},
	"040-add-space-id-wi-wit-tq.sql": &_bintree_t{_040_add_space_id_wi_wit_tq_sql, map[string]*_bintree_t{
	}},
	"041-unique-area-name-create-new-area.sql": &_bintree_t{_041_unique_area_name_create_new_area_sql, map[string]*_bintree_t{
	}},
	"042-work-item-link-revisions.sql": &_bintree_t{_042_work_item_link_revisions_sql, map[string]*_bintree_t{
	}},
	"043-space-resources.sql": &_bintree_t{_043_space_resources_sql, map[string]*_bintree_t{
	}},
	"044-add-contextinfo-column-users.sql": &_bintree_t{_044_add_contextinfo_column_users_sql, map[string]*_bintree_t{
	}},
	"045-adds-order-to-existing-wi.sql": &_bintree_t{_045_adds_order_to_existing_wi_sql, map[string]*_bintree_t{
	}},
	"046-oauth-states.sql": &_bintree_t{_046_oauth_states_sql, map[string]*_bintree_t{
	}},
	"047-codebases.sql": &_bintree_t{_047_codebases_sql, map[string]*_bintree_t{
	}},
	"048-unique-iteration-name-create-new-iteration.sql": &_bintree_t{_048_unique_iteration_name_create_new_iteration_sql, map[string]*_bintree_t{
	}},
	"049-add-wi-to-root-area.sql": &_bintree_t{_049_add_wi_to_root_area_sql, map[string]*_bintree_t{
	}},
	"050-add-company-to-user-profile.sql": &_bintree_t{_050_add_company_to_user_profile_sql, map[string]*_bintree_t{
	}},
	"051-modify-work_item_link_types_name_idx.sql": &_bintree_t{_051_modify_work_item_link_types_name_idx_sql, map[string]*_bintree_t{
	}},
	"052-unique-space-names.sql": &_bintree_t{_052_unique_space_names_sql, map[string]*_bintree_t{
	}},
	"053-edit-username.sql": &_bintree_t{_053_edit_username_sql, map[string]*_bintree_t{
	}},
	"054-add-stackid-to-codebase.sql": &_bintree_t{_054_add_stackid_to_codebase_sql, map[string]*_bintree_t{
	}},
	"055-assign-root-area-if-missing.sql": &_bintree_t{_055_assign_root_area_if_missing_sql, map[string]*_bintree_t{
	}},
	"056-assign-root-iteration-if-missing.sql": &_bintree_t{_056_assign_root_iteration_if_missing_sql, map[string]*_bintree_t{
	}},
	"057-add-last-used-workspace-to-codebase.sql": &_bintree_t{_057_add_last_used_workspace_to_codebase_sql, map[string]*_bintree_t{
	}},
	"058-index-identities-fullname.sql": &_bintree_t{_058_index_identities_fullname_sql, map[string]*_bintree_t{
	}},
	"059-fixed-ids-for-system-link-types-and-categories.sql": &_bintree_t{_059_fixed_ids_for_system_link_types_and_categories_sql, map[string]*_bintree_t{
	}},
	"060-fixed-identities-username-idx.sql": &_bintree_t{_060_fixed_identities_username_idx_sql, map[string]*_bintree_t{
	}},
	"061-replace-index-space-name.sql": &_bintree_t{_061_replace_index_space_name_sql, map[string]*_bintree_t{
	}},
	"062-link-system-preparation.sql": &_bintree_t{_062_link_system_preparation_sql, map[string]*_bintree_t{
	}},
	"063-workitem-related-changes.sql": &_bintree_t{_063_workitem_related_changes_sql, map[string]*_bintree_t{
	}},
	"064-remove-link-combinations.sql": &_bintree_t{_064_remove_link_combinations_sql, map[string]*_bintree_t{
	}},
	"065-workitem-id-unique-per-space.sql": &_bintree_t{_065_workitem_id_unique_per_space_sql, map[string]*_bintree_t{
	}},
	"066-work_item_links_data_integrity.sql": &_bintree_t{_066_work_item_links_data_integrity_sql, map[string]*_bintree_t{
	}},
	"067-comment-parentid-uuid.sql": &_bintree_t{_067_comment_parentid_uuid_sql, map[string]*_bintree_t{
	}},
	"068-index_identities_username.sql": &_bintree_t{_068_index_identities_username_sql, map[string]*_bintree_t{
	}},
	"069-limit-execution-order-to-space.sql": &_bintree_t{_069_limit_execution_order_to_space_sql, map[string]*_bintree_t{
	}},
	"070-rename-comment-createdby-to-creator.sql": &_bintree_t{_070_rename_comment_createdby_to_creator_sql, map[string]*_bintree_t{
	}},
	"071-iteration-related-changes.sql": &_bintree_t{_071_iteration_related_changes_sql, map[string]*_bintree_t{
	}},
	"072-adds-active-flag-in-iteration.sql": &_bintree_t{_072_adds_active_flag_in_iteration_sql, map[string]*_bintree_t{
	}},
	"073-labels.sql": &_bintree_t{_073_labels_sql, map[string]*_bintree_t{
	}},
	"074-label-border-color.sql": &_bintree_t{_074_label_border_color_sql, map[string]*_bintree_t{
	}},
	"075-label-unique-name.sql": &_bintree_t{_075_label_unique_name_sql, map[string]*_bintree_t{
	}},
	"076-drop-space-resources-and-oauth-state.sql": &_bintree_t{_076_drop_space_resources_and_oauth_state_sql, map[string]*_bintree_t{
	}},
	"077-index-work-item-links.sql": &_bintree_t{_077_index_work_item_links_sql, map[string]*_bintree_t{
	}},
	"078-assignee-and-label-empty-value.sql": &_bintree_t{_078_assignee_and_label_empty_value_sql, map[string]*_bintree_t{
	}},
	"078-tracker-to-use-uuid.sql": &_bintree_t{_078_tracker_to_use_uuid_sql, map[string]*_bintree_t{
	}},
	"079-assignee-and-label-empty-value.sql": &_bintree_t{_079_assignee_and_label_empty_value_sql, map[string]*_bintree_t{
	}},
	"080-remove-unknown-link-types.sql": &_bintree_t{_080_remove_unknown_link_types_sql, map[string]*_bintree_t{
	}},
	"081-queries.sql": &_bintree_t{_081_queries_sql, map[string]*_bintree_t{
	}},
	"082-iteration-related-changes.sql": &_bintree_t{_082_iteration_related_changes_sql, map[string]*_bintree_t{
	}},
	"083-index-comments-parent.sql": &_bintree_t{_083_index_comments_parent_sql, map[string]*_bintree_t{
	}},
	"084-codebases-spaceid-url-index.sql": &_bintree_t{_084_codebases_spaceid_url_index_sql, map[string]*_bintree_t{
	}},
	"085-delete-system.number-json-field.sql": &_bintree_t{_085_delete_system_number_json_field_sql, map[string]*_bintree_t{
	}},
	"086-add-can-construct-to-wit.sql": &_bintree_t{_086_add_can_construct_to_wit_sql, map[string]*_bintree_t{
	}},
	"087-space-templates.sql": &_bintree_t{_087_space_templates_sql, map[string]*_bintree_t{
	}},
	"088-type-groups-and-child-types.sql": &_bintree_t{_088_type_groups_and_child_types_sql, map[string]*_bintree_t{
	}},
	"089-fixup-space-templates.sql": &_bintree_t{_089_fixup_space_templates_sql, map[string]*_bintree_t{
	}},
	"090-queries-version.sql": &_bintree_t{_090_queries_version_sql, map[string]*_bintree_t{
	}},
	"091-comments-child-comments.sql": &_bintree_t{_091_comments_child_comments_sql, map[string]*_bintree_t{
	}},
	"092-comment-revisions-child-comments.sql": &_bintree_t{_092_comment_revisions_child_comments_sql, map[string]*_bintree_t{
	}},
	"093-codebase-add-cve-scan.sql": &_bintree_t{_093_codebase_add_cve_scan_sql, map[string]*_bintree_t{
	}},
	"094-changes-to-agile-template.sql": &_bintree_t{_094_changes_to_agile_template_sql, map[string]*_bintree_t{
	}},
	"095-remove-resolution-field-from-impediment.sql": &_bintree_t{_095_remove_resolution_field_from_impediment_sql, map[string]*_bintree_t{
	}},
	"096-changes-to-agile-template.sql": &_bintree_t{_096_changes_to_agile_template_sql, map[string]*_bintree_t{
	}},
	"097-remove-resolution-field-from-impediment.sql": &_bintree_t{_097_remove_resolution_field_from_impediment_sql, map[string]*_bintree_t{
	}},
	"README.adoc": &_bintree_t{readme_adoc, map[string]*_bintree_t{
	}},
}}
