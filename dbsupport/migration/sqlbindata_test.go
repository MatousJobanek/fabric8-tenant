package migration_test

import (
	"fmt"
	"strings"
)
var __044_insert_test_data_sql = []byte(`-- users
INSERT INTO
   users(created_at, updated_at, id, email, full_name, image_url, bio, url, context_information)
VALUES
   (
      now(), now(), '01b291cd-9399-4f1a-8bbc-d1de66d76192', 'testone@example.com', 'test one', 'https://www.gravatar.com/avatar/testone', 'my test bio one', 'http://example.com', '{"key": "value"}'
   ),
   (
      now(), now(), '0d19928e-ef61-46fd-9bdc-71d1ecbce2c7', 'testtwo@example.com', 'test two', 'http://https://www.gravatar.com/avatar/testtwo', 'my test bio two', 'http://example.com', '{"key": "value"}'
   )
;
-- identities
INSERT INTO
   identities(created_at, updated_at, id, username, provider_type, user_id, profile_url)
VALUES
   (
      now(), now(), '01b291cd-9399-4f1a-8bbc-d1de66d76192', 'testone', 'github', '01b291cd-9399-4f1a-8bbc-d1de66d76192', 'http://example-github.com'
   ),
   (
      now(), now(), '5f946975-ff47-4c4a-b5dc-778f0b7e476c', 'testwo', 'rhhd', '0d19928e-ef61-46fd-9bdc-71d1ecbce2c7', 'http://example-rhd.com'
   )
;
-- spaces
INSERT INTO
   spaces (created_at, updated_at, id, version, name, description, owner_id)
VALUES
   (
      now(), now(), '86af5178-9b41-469b-9096-57e5155c3f31', 0, 'test.space.one', 'space desc one', '01b291cd-9399-4f1a-8bbc-d1de66d76192'
   )
;
-- work_item_types
INSERT INTO
   work_item_types(created_at, updated_at, id, name, version, fields, space_id)
VALUES
   (
      now(), now(), 'bbf35418-04b6-426c-a60b-7f80beb0b624', 'Test item type 1', 1.0, '{}', '2e0698d8-753e-4cef-bb7c-f027634824a2'
   )
;
INSERT INTO
   work_item_types(created_at, updated_at, id, name, version, path, fields, space_id)
VALUES
   (
      now(), now(), '86af5178-9b41-469b-9096-57e5155c3f31', 'Test item type 2', 1.0, 'bbf35418_04b6_426c_a60b_7f80beb0b624.86af5178_9b41_469b_9096_57e5155c3f31', '{}', '86af5178-9b41-469b-9096-57e5155c3f31'
   )
;
-- trackers
INSERT INTO
   trackers(created_at, updated_at, id, url, type)
VALUES
   (
      now(), now(), 1, 'http://example.com', 'github'
   ),
   (
      now(), now(), 2, 'http://example-jira.com', 'jira'
   )
;
-- tracker_queries id | query | schedule | tracker_id | space_id
INSERT INTO
   tracker_queries(created_at, updated_at, id, query, schedule, tracker_id, space_id)
VALUES
   (
      now(), now(), 1, 'SELECT * FROM', 'schedule', 1, '86af5178-9b41-469b-9096-57e5155c3f31'
   ),
   (
      now(), now(), 2, 'SELECT * FROM', 'schedule', 2, '86af5178-9b41-469b-9096-57e5155c3f31'
   )
;

-- space_resources
INSERT INTO
   space_resources(created_at, updated_at, id, space_id, resource_id, policy_id, permission_id)
VALUES
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', '86af5178-9b41-469b-9096-57e5155c3f31', 'resource_id', 'policy_id', 'permission_id'
   ),
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', '86af5178-9b41-469b-9096-57e5155c3f31', 'resource_id', 'policy_id', 'permission_id'
   )
;
-- areas created_at | updated_at | deleted_at | id | space_id | version | path | name
INSERT INTO
   areas(created_at, updated_at, id, space_id, version, path, name)
VALUES
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', '86af5178-9b41-469b-9096-57e5155c3f31', 0, 'path', 'area test one'
   ),
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', '86af5178-9b41-469b-9096-57e5155c3f31', 0, '', 'area test two'
   )
;
-- iterations
INSERT INTO
   iterations(created_at, updated_at, id, space_id, start_at, end_at, name, description, state)
VALUES
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', '86af5178-9b41-469b-9096-57e5155c3f31', now(), now(), 'iteration test one', 'description', 'new'
   ),
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', '86af5178-9b41-469b-9096-57e5155c3f31', now(), now(), 'iteration test two', 'description', 'start'
   )
;
-- comments
INSERT INTO
   comments(created_at, updated_at, id, parent_id, body, created_by, markup)
VALUES
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', '2e0698d8-753e-4cef-bb7c-f027634824a2', 'body test one', '01b291cd-9399-4f1a-8bbc-d1de66d76192', 'PlainText'
   ),
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', '2e0698d8-753e-4cef-bb7c-f027634824a2', 'body test two', '01b291cd-9399-4f1a-8bbc-d1de66d76192', 'PlainText'
   )
;
-- comment_revisions
INSERT INTO
   comment_revisions(id, revision_time, revision_type, modifier_id, comment_id, comment_body, comment_markup, comment_parent_id)
VALUES
   (
      '71171e90-6d35-498f-a6a7-2083b5267c18', now(), 1, '5f946975-ff47-4c4a-b5dc-778f0b7e476c', '71171e90-6d35-498f-a6a7-2083b5267c18', 'comment body test one', 'comment markup test one', '71171e90-6d35-498f-a6a7-2083b5267c18'
   ),
   (
      '2e0698d8-753e-4cef-bb7c-f027634824a2', now(), 1, '5f946975-ff47-4c4a-b5dc-778f0b7e476c', '71171e90-6d35-498f-a6a7-2083b5267c18', 'comment body test two', 'comment markup test two', '71171e90-6d35-498f-a6a7-2083b5267c18'
   )
;
-- work_item_link_categories
INSERT INTO
   work_item_link_categories(created_at, updated_at, id, version, name, description)
VALUES
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', 1, 'name test one', 'description one'
   ),
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', 1, 'name test two', 'description two'
   )
;
-- work_item_link_types
INSERT INTO
   work_item_link_types(created_at, updated_at, id, version, name, description, forward_name, reverse_name, topology, link_category_id, space_id, source_type_id, target_type_id)
VALUES
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', 1, 'test one', 'desc', 'forward test one', 'reverser test one', 'dependency', '71171e90-6d35-498f-a6a7-2083b5267c18', '2e0698d8-753e-4cef-bb7c-f027634824a2', '86af5178-9b41-469b-9096-57e5155c3f31', '86af5178-9b41-469b-9096-57e5155c3f31'
   ),
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', 1, 'test two', 'desc', 'forward test two', 'reverser test two', 'network', '2e0698d8-753e-4cef-bb7c-f027634824a2', '2e0698d8-753e-4cef-bb7c-f027634824a2', '86af5178-9b41-469b-9096-57e5155c3f31', '86af5178-9b41-469b-9096-57e5155c3f31'
   )
;
-- work_items
INSERT INTO
   work_items(created_at, updated_at, type, version, space_id, fields)
VALUES
   (
      now(), now(), 'bbf35418-04b6-426c-a60b-7f80beb0b624', 1.0, '86af5178-9b41-469b-9096-57e5155c3f31', '{}'
   ),
   (
      now(), now(), 'bbf35418-04b6-426c-a60b-7f80beb0b624', 2.0, '86af5178-9b41-469b-9096-57e5155c3f31', '{}'
   )
;
-- work_item_revisions
INSERT INTO
   work_item_revisions(id, revision_time, revision_type, modifier_id, work_item_id, work_item_type_id, work_item_version, work_item_fields)
VALUES
   (
      '2e0698d8-753e-4cef-bb7c-f027634824a2', now(), 1, '01b291cd-9399-4f1a-8bbc-d1de66d76192', 1, '2e0698d8-753e-4cef-bb7c-f027634824a2', 1, '{}'
   ),
   (
      '71171e90-6d35-498f-a6a7-2083b5267c18', now(), 1, '01b291cd-9399-4f1a-8bbc-d1de66d76192', 1, '2e0698d8-753e-4cef-bb7c-f027634824a2', 1, '{}'
   )
;
-- work_item_links
INSERT INTO
   work_item_links(created_at, updated_at, id, version, link_type_id)
VALUES
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', 1, '2e0698d8-753e-4cef-bb7c-f027634824a2'
   ),
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', 1, '71171e90-6d35-498f-a6a7-2083b5267c18'
   )
;
-- work_item_link_revisions
INSERT INTO
   work_item_link_revisions(id, revision_time, revision_type, modifier_id, work_item_link_id, work_item_link_version, work_item_link_source_id, work_item_link_target_id, work_item_link_type_id)
VALUES
   (
      '71171e90-6d35-498f-a6a7-2083b5267c18', now(), 1, '01b291cd-9399-4f1a-8bbc-d1de66d76192', '71171e90-6d35-498f-a6a7-2083b5267c18', 1, 1, 2, '2e0698d8-753e-4cef-bb7c-f027634824a2'
   ),
   (
      '2e0698d8-753e-4cef-bb7c-f027634824a2', now(), 2, '01b291cd-9399-4f1a-8bbc-d1de66d76192', '71171e90-6d35-498f-a6a7-2083b5267c18', 1, 2, 1, '2e0698d8-753e-4cef-bb7c-f027634824a2'
   )
;
-- tracker_items
INSERT INTO
   tracker_items(created_at, updated_at, id, remote_item_id, item, batch_id, tracker_id)
VALUES
   (
      now(), now(), 1, 'remote_id', 'test one', 'batch_id', 1
   ),
   (
      now(), now(), 2, 'remote_id', 'test two', 'batch_id', 2
   )
;
`)

func _044_insert_test_data_sql() ([]byte, error) {
	return __044_insert_test_data_sql, nil
}

var __045_update_work_items_sql = []byte(`-- work_items
UPDATE work_items SET execution_order=100000 WHERE type='bbf35418-04b6-426c-a60b-7f80beb0b624';
UPDATE work_items SET execution_order=200000 WHERE type='bbf35418-04b6-426c-a60b-7f80beb0b624';
`)

func _045_update_work_items_sql() ([]byte, error) {
	return __045_update_work_items_sql, nil
}

var __046_insert_oauth_states_sql = []byte(`-- oauth_state_references
INSERT INTO
   oauth_state_references(created_at, updated_at, id, referrer)
VALUES
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', 'test referrer one text'
   )
;
INSERT INTO
   oauth_state_references(created_at, updated_at, id, referrer)
VALUES
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', 'test referrer two text'
   )
;
`)

func _046_insert_oauth_states_sql() ([]byte, error) {
	return __046_insert_oauth_states_sql, nil
}

var __047_insert_codebases_sql = []byte(`-- codebases
INSERT INTO
   codebases(created_at, updated_at, id, space_id, type, url)
VALUES
   (
      now(), now(), '2e0698d8-753e-4cef-bb7c-f027634824a2', '86af5178-9b41-469b-9096-57e5155c3f31', 'type test one', 'http://example-jira.com'
   )
;
INSERT INTO
   codebases(created_at, updated_at, id, space_id, type, url)
VALUES
   (
      now(), now(), '71171e90-6d35-498f-a6a7-2083b5267c18', '86af5178-9b41-469b-9096-57e5155c3f31', 'type test two', 'http://example-jira.com'
   )
;
`)

func _047_insert_codebases_sql() ([]byte, error) {
	return __047_insert_codebases_sql, nil
}

var __048_unique_idx_failed_insert_sql = []byte(`-- insert two iterations one will fail due to invalid iterations_name_space_id_path_unique
INSERT INTO
   iterations(created_at, updated_at, id, space_id, start_at, end_at, name, description, state, "path")
VALUES
   (
      now(), now(), '86af5178-9b41-469b-9096-57e5155c3f31', '86af5178-9b41-469b-9096-57e5155c3f31', now(), now(), 'iteration test one', 'description', 'new', '/'
   )
;

INSERT INTO
   iterations(created_at, updated_at, id, space_id, start_at, end_at, name, description, state, "path")
VALUES
   (
      now(), now(), '0a24d3c2-e0a6-4686-8051-ec0ea1915a28', '86af5178-9b41-469b-9096-57e5155c3f31', now(), now(), 'iteration test one', 'description', 'new', '/'
   )
;
`)

func _048_unique_idx_failed_insert_sql() ([]byte, error) {
	return __048_unique_idx_failed_insert_sql, nil
}

var __050_users_add_column_company_sql = []byte(`-- Set company value to the existing users
UPDATE users SET company='RedHat Inc.' WHERE full_name='test one' OR full_name='test two';
`)

func _050_users_add_column_company_sql() ([]byte, error) {
	return __050_users_add_column_company_sql, nil
}

var __053_edit_username_sql = []byte(`-- users
INSERT INTO
   users(created_at, updated_at, id, email, full_name, image_url, bio, url, context_information)
VALUES
   (
      now(), now(), 'f03f023b-0427-4cdb-924b-fb2369018ab7', 'test2@example.com', 'test1', 'https://www.gravatar.com/avatar/testtwo2', 'my test bio one', 'http://example.com/001', '{"key": "value"}'
   ),
   (
      now(), now(), 'f03f023b-0427-4cdb-924b-fb2369018ab6', 'test3@example.com', 'test2', 'http://https://www.gravatar.com/avatar/testtwo3', 'my test bio two', 'http://example.com/002', '{"key": "value"}'
   )
;
-- identities
INSERT INTO
   identities(created_at, updated_at, id, username, provider_type, user_id, profile_url)
VALUES
   (
      now(), now(), '2a808366-9525-4646-9c80-ed704b2eebbe', 'test1', 'github', 'f03f023b-0427-4cdb-924b-fb2369018ab7', 'http://example-github.com/001'
   ),
   (
      now(), now(), '2a808366-9525-4646-9c80-ed704b2eebbb', 'test2', 'rhhd', 'f03f023b-0427-4cdb-924b-fb2369018ab6', 'http://example-rhd.com/002'
   )
;
`)

func _053_edit_username_sql() ([]byte, error) {
	return __053_edit_username_sql, nil
}

var __054_add_stackid_to_codebase_sql = []byte(`UPDATE codebases set stack_id ='java-centos';
`)

func _054_add_stackid_to_codebase_sql() ([]byte, error) {
	return __054_add_stackid_to_codebase_sql, nil
}

var __055_assign_root_area_if_missing_sql = []byte(`insert into spaces (id, name) values ('11111111-2222-0000-0000-000000000000', 'test');
insert into areas (id, name, path, space_id) values ('11111111-3333-0000-0000-000000000000', 'test area', '', '11111111-2222-0000-0000-000000000000');
insert into work_item_types (id, name, space_id) values ('11111111-4444-0000-0000-000000000000', 'Test WIT','11111111-2222-0000-0000-000000000000');
insert into work_items (id, space_id, type, fields) values (12345, '11111111-2222-0000-0000-000000000000', '11111111-4444-0000-0000-000000000000', '{"system.title":"Title"}'::json);`)

func _055_assign_root_area_if_missing_sql() ([]byte, error) {
	return __055_assign_root_area_if_missing_sql, nil
}

var __056_assign_root_iteration_if_missing_sql = []byte(`insert into spaces (id, name) values ('11111111-2222-bbbb-0000-000000000000', 'test');
insert into iterations (id, name, path, space_id) values ('11111111-3333-bbbb-0000-000000000000', 'test area', '', '11111111-2222-bbbb-0000-000000000000');
insert into work_item_types (id, name, space_id) values ('11111111-4444-bbbb-0000-000000000000', 'Test WIT','11111111-2222-bbbb-0000-000000000000');
insert into work_items (id, space_id, type, fields) values (12346, '11111111-2222-bbbb-0000-000000000000', '11111111-4444-bbbb-0000-000000000000', '{"system.title":"Title"}'::json);`)

func _056_assign_root_iteration_if_missing_sql() ([]byte, error) {
	return __056_assign_root_iteration_if_missing_sql, nil
}

var __057_add_last_used_workspace_to_codebase_sql = []byte(`UPDATE codebases set last_used_workspace ='java-centos-last-workspace';
`)

func _057_add_last_used_workspace_to_codebase_sql() ([]byte, error) {
	return __057_add_last_used_workspace_to_codebase_sql, nil
}

var __061_add_duplicate_space_owner_name_sql = []byte(`--- added a duplicate space with the same owner and name than a previous one
INSERT INTO
   spaces (created_at, updated_at, id, version, name, description, owner_id)
VALUES
   (
      now(), now(), '86af5178-9b41-469b-9096-57e5155c3f32', 0, 'test.Space.one', 'Space desc one', '01b291cd-9399-4f1a-8bbc-d1de66d76192'
   )
;
`)

func _061_add_duplicate_space_owner_name_sql() ([]byte, error) {
	return __061_add_duplicate_space_owner_name_sql, nil
}

var __063_workitem_related_changes_sql = []byte(`--
-- comments
--
insert into spaces (id, name) values ('11111111-6262-0000-0000-000000000000', 'test');
insert into work_item_types (id, name, space_id) values ('11111111-6262-0000-0000-000000000000', 'Test WIT','11111111-6262-0000-0000-000000000000');
insert into work_items (id, space_id, type, fields) values (62001, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 1"}'::json);
insert into work_items (id, space_id, type, fields) values (62002, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 2"}'::json);
insert into work_items (id, space_id, type, fields) values (62003, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
-- remove previous comments
delete from comments;
-- add comments linked to work items above
insert into comments (id, parent_id, body, created_at) values ( '11111111-6262-0001-0000-000000000000', '62001', 'a comment', '2017-06-13 09:00:00.0000+00');
insert into comments (id, parent_id, body, created_at) values ( '11111111-6262-0003-0000-000000000000', '62003', 'a comment', '2017-06-13 11:00:00.0000+00');
update comments set deleted_at = '2017-06-13 11:15:00.0000+00' where id =  '11111111-6262-0003-0000-000000000000';

--
-- work item links
--
insert into work_items (id, space_id, type, fields) values (62004, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, space_id, type, fields) values (62005, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, space_id, type, fields) values (62006, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, space_id, type, fields) values (62007, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, space_id, type, fields) values (62008, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, space_id, type, fields) values (62009, '11111111-6262-0000-0000-000000000000', '11111111-6262-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
delete from work_item_links;
insert into work_item_links (id, version, source_id, target_id, created_at) values ('11111111-6262-0001-0000-000000000000', 1, 62004, 62005, '2017-06-13 09:00:00.0000+00');
insert into work_item_links (id, version, source_id, target_id, deleted_at) values ('11111111-6262-0003-0000-000000000000', 1, 62008, 62009, '2017-06-13 11:00:00.0000+00');
update work_item_links set deleted_at = '2017-06-13 11:15:00.0000+00' where id = '11111111-6262-0003-0000-000000000000';


`)

func _063_workitem_related_changes_sql() ([]byte, error) {
	return __063_workitem_related_changes_sql, nil
}

var __065_workitem_id_unique_per_space_sql = []byte(`-- create spaces 1 and 2
insert into spaces (id, name) values ('11111111-0000-0000-0000-000000000000', 'test space 1');
insert into spaces (id, name) values ('22222222-0000-0000-0000-000000000000', 'test space 2');
-- create work item types for spaces 1 and 2
insert into work_item_types (id, name, space_id) values ('11111111-0000-0000-0000-000000000000', 'test type 1', '11111111-0000-0000-0000-000000000000');
insert into work_item_types (id, name, space_id) values ('22222222-0000-0000-0000-000000000000', 'test type 2', '22222222-0000-0000-0000-000000000000');
-- create work item link types for spaces 1 and 2
insert into work_item_link_types (id, name, topology, forward_name, reverse_name, space_id) 
    values ('11111111-0000-0000-0000-000000000000', 'foo', 'dependency', 'foo', 'foo', '11111111-0000-0000-0000-000000000000');
insert into work_item_link_types (id, name, topology, forward_name, reverse_name, space_id) 
    values ('22222222-0000-0000-0000-000000000000', 'bar', 'dependency', 'bar', 'bar', '22222222-0000-0000-0000-000000000000');
-- create identity (for revisions)
insert into identities (id, username) values ('cafebabe-0000-0000-0000-000000000000', 'foo');
-- inserting work items, their revisions and comments in space '1'
insert into work_items (id, type, space_id) values (12347, '11111111-0000-0000-0000-000000000000', '11111111-0000-0000-0000-000000000000');
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (1, 'cafebabe-0000-0000-0000-000000000000', 12347);
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (2, 'cafebabe-0000-0000-0000-000000000000', 12347);
insert into comments (parent_id, body) values ('12347', 'blabla');
insert into work_items (id, type, space_id) values (12348, '11111111-0000-0000-0000-000000000000', '11111111-0000-0000-0000-000000000000');
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (1, 'cafebabe-0000-0000-0000-000000000000', 12348);
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (2, 'cafebabe-0000-0000-0000-000000000000', 12348);
insert into comments (parent_id, body) values ('12348', 'blabla');
-- inserting work items, their revisions and comments in space '2'
insert into work_items (id, type, space_id) values (12349, '22222222-0000-0000-0000-000000000000', '22222222-0000-0000-0000-000000000000');
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (1, 'cafebabe-0000-0000-0000-000000000000', 12349);
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (2, 'cafebabe-0000-0000-0000-000000000000', 12349);
insert into comments (parent_id, body) values ('12349', 'blabla');
insert into work_items (id, type, space_id) values (12350, '22222222-0000-0000-0000-000000000000', '22222222-0000-0000-0000-000000000000');
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (1, 'cafebabe-0000-0000-0000-000000000000', 12350);
insert into work_item_revisions (revision_type, modifier_id, work_item_id) values (2, 'cafebabe-0000-0000-0000-000000000000', 12350);
insert into comments (parent_id, body) values ('12350', 'blabla');
-- insert links between work items
insert into work_item_links (id, link_type_id, source_id, target_id) values ('11111111-0000-0000-0000-000000000000', '11111111-0000-0000-0000-000000000000', 12347, 12348);
insert into work_item_link_revisions (revision_type, modifier_id, work_item_link_id, work_item_link_version, work_item_link_source_id, work_item_link_target_id, work_item_link_type_id)
  values (1, 'cafebabe-0000-0000-0000-000000000000', '11111111-0000-0000-0000-000000000000',0,12347,12348,'11111111-0000-0000-0000-000000000000');
insert into work_item_links (id, link_type_id, source_id, target_id) values ('22222222-0000-0000-0000-000000000000', '22222222-0000-0000-0000-000000000000', 12349, 12350);
insert into work_item_link_revisions (revision_type, modifier_id, work_item_link_id, work_item_link_version, work_item_link_source_id, work_item_link_target_id, work_item_link_type_id)
  values (1, 'cafebabe-0000-0000-0000-000000000000', '22222222-0000-0000-0000-000000000000',0,12349,12350,'22222222-0000-0000-0000-000000000000');
`)

func _065_workitem_id_unique_per_space_sql() ([]byte, error) {
	return __065_workitem_id_unique_per_space_sql, nil
}

var __066_work_item_links_data_integrity_sql = []byte(`-- prepare data
insert into spaces (id, name) values ('00000066-0000-0000-0000-000000000000', 'test space 1');
insert into work_item_types (id, name, space_id) values ('00000066-0000-0000-0000-000000000000', 'test type 1', '00000066-0000-0000-0000-000000000000');
insert into work_item_link_types (id, name, topology, forward_name, reverse_name, space_id) 
    values ('00000066-0000-0000-0000-000000000000', 'foo', 'dependency', 'foo', 'foo', '00000066-0000-0000-0000-000000000000');
insert into work_items (id, type, space_id) values ('00000066-0000-0000-0000-000000000001', '00000066-0000-0000-0000-000000000000', '00000066-0000-0000-0000-000000000000');
insert into work_items (id, type, space_id) values ('00000066-0000-0000-0000-000000000002', '00000066-0000-0000-0000-000000000000', '00000066-0000-0000-0000-000000000000');
-- insert valid and invalid links
insert into work_item_links (id, link_type_id, source_id, target_id) values ('00000066-0000-0000-0000-000000000001', '00000066-0000-0000-0000-000000000000', '00000066-0000-0000-0000-000000000001', '00000066-0000-0000-0000-000000000002');
insert into work_item_links (id, link_type_id, source_id, target_id) values ('00000066-0000-0000-0000-000000000002', NULL, '00000066-0000-0000-0000-000000000001', '00000066-0000-0000-0000-000000000002');
insert into work_item_links (id, link_type_id, source_id, target_id) values ('00000066-0000-0000-0000-000000000003', '00000066-0000-0000-0000-000000000000', NULL, '00000066-0000-0000-0000-000000000002');
insert into work_item_links (id, link_type_id, source_id, target_id) values ('00000066-0000-0000-0000-000000000004', '00000066-0000-0000-0000-000000000000', '00000066-0000-0000-0000-000000000001', NULL);
`)

func _066_work_item_links_data_integrity_sql() ([]byte, error) {
	return __066_work_item_links_data_integrity_sql, nil
}

var __067_comment_parentid_uuid_sql = []byte(`-- need some work items to migrate the comment_revisions table, too
insert into spaces (id, name) values ('00000067-0000-0000-0000-000000000000', 'test space 67');
insert into work_item_types (id, name, space_id) values ('00000067-0000-0000-0000-000000000000', 'test type 1', '00000067-0000-0000-0000-000000000000');
insert into work_items (id, number, type, space_id) values ('00000067-0000-0000-0000-000000000000', 1, '00000067-0000-0000-0000-000000000000', '00000067-0000-0000-0000-000000000000');
insert into comments (id, parent_id, body) values ('00000067-0000-0000-0000-000000000000', '00000067-0000-0000-0000-000000000000', 'a foo comment');
insert into comment_revisions (id, revision_type, modifier_id, comment_id, comment_parent_id, comment_body) 
    values ('00000067-0000-0000-0000-000000000000', 1, 'cafebabe-0000-0000-0000-000000000000', '00000067-0000-0000-0000-000000000000',  1, 'a foo comment');`)

func _067_comment_parentid_uuid_sql() ([]byte, error) {
	return __067_comment_parentid_uuid_sql, nil
}

var __071_iteration_related_changes_sql = []byte(`insert into spaces (id, name) values ('11111111-7171-0000-0000-000000000000', 'test iteration - relationships changed at');
delete from work_items;
insert into work_item_types (id, name, space_id) values ('11111111-7171-0000-0000-000000000000', 'Test WIT','11111111-7171-0000-0000-000000000000');
insert into work_items (id, created_at, space_id, type, fields) values ('11111111-7171-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-7171-0000-0000-000000000000', '11111111-7171-0000-0000-000000000000', '{"system.title":"Work item 1"}'::json);
insert into work_items (id, created_at, space_id, type, fields) values ('22222222-7171-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-7171-0000-0000-000000000000', '11111111-7171-0000-0000-000000000000', '{"system.title":"Work item 2"}'::json);
insert into work_items (id, created_at, space_id, type, fields) values ('33333333-7171-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-7171-0000-0000-000000000000', '11111111-7171-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, created_at, space_id, type, fields) values ('44444444-7171-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-7171-0000-0000-000000000000', '11111111-7171-0000-0000-000000000000', '{"system.title":"Work item 4"}'::json);

delete from iterations;
insert into iterations (id, name, created_at, space_id) values ('11111111-7171-0000-0000-000000000000', 'iteration 1', CURRENT_TIMESTAMP, '11111111-7171-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('22222222-7171-0000-0000-000000000000', 'iteration 2', CURRENT_TIMESTAMP, '11111111-7171-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('33333333-7171-0000-0000-000000000000', 'iteration 3', CURRENT_TIMESTAMP, '11111111-7171-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('44444444-7171-0000-0000-000000000000', 'iteration 4', CURRENT_TIMESTAMP, '11111111-7171-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('55555555-7171-0000-0000-000000000000', 'iteration 5', CURRENT_TIMESTAMP, '11111111-7171-0000-0000-000000000000');

-- link work item 1 to iteration 1
update work_items set updated_at = (CURRENT_TIMESTAMP + interval '1 hour'), fields = '{"system.title":"Work item 1", "system.iteration":"11111111-7171-0000-0000-000000000000"}'::json where id = '11111111-7171-0000-0000-000000000000';
-- link work item 2 to iteration 2 then iteration 3
update work_items set updated_at = (CURRENT_TIMESTAMP + interval '1 hour'), fields = '{"system.title":"Work item 2", "system.iteration":"22222222-7171-0000-0000-000000000000"}'::json where id = '22222222-7171-0000-0000-000000000000';
update work_items set updated_at = (CURRENT_TIMESTAMP + interval '2 hour'), fields = '{"system.title":"Work item 2", "system.iteration":"33333333-7171-0000-0000-000000000000"}'::json where id = '22222222-7171-0000-0000-000000000000';
-- link work item 3 to iteration 4 then soft-delete the work item
update work_items set fields = '{"system.title":"Work item 3", "system.iteration":"44444444-7171-0000-0000-000000000000"}'::json, updated_at = (CURRENT_TIMESTAMP + interval '1 hour') where id = '33333333-7171-0000-0000-000000000000';
update work_items set deleted_at = (CURRENT_TIMESTAMP + interval '2 hour') where id = '33333333-7171-0000-0000-000000000000';
-- link work item 4 to iteration 5 then set another, unrelated field
update work_items set fields = '{"system.title":"Work item 4", "system.iteration":"55555555-7171-0000-0000-000000000000"}'::json, updated_at = (CURRENT_TIMESTAMP + interval '1 hour') where id = '44444444-7171-0000-0000-000000000000';
update work_items set fields = '{"system.title":"Work item 4", "system.iteration":"55555555-7171-0000-0000-000000000000", "system.description":"foo"}'::json, updated_at = (CURRENT_TIMESTAMP + interval '2 hour') where id = '44444444-7171-0000-0000-000000000000';

`)

func _071_iteration_related_changes_sql() ([]byte, error) {
	return __071_iteration_related_changes_sql, nil
}

var __073_label_color_code_sql = []byte(`-- wrong color code
-- this should fail
DELETE FROM labels;
DELETE FROM spaces where id='11111111-7171-0000-0000-000000000000';
INSERt INTO spaces (id, name) VALUES ('11111111-7171-0000-0000-000000000000', 'test space');
INSERT INTO labels (name, text_color, background_color, space_id) VALUES ('some name', '#rrsstt', '#2f4c56', '11111111-7171-0000-0000-000000000000');


`)

func _073_label_color_code_sql() ([]byte, error) {
	return __073_label_color_code_sql, nil
}

var __073_label_color_code2_sql = []byte(`-- wrong color code
-- this should fail
DELETE FROM labels;
DELETE FROM spaces where id='11111111-7171-0000-0000-000000000000';
INSERt INTO spaces (id, name) VALUES ('11111111-7171-0000-0000-000000000000', 'test space');
INSERT INTO labels (name, text_color, background_color, space_id) VALUES ('some name', '#2f4c56', '#rrsstt', '11111111-7171-0000-0000-000000000000');
`)

func _073_label_color_code2_sql() ([]byte, error) {
	return __073_label_color_code2_sql, nil
}

var __073_label_empty_name_sql = []byte(`-- empty label name
-- this should fail
delete from labels;
INSERT INTO labels(text_color, background_color, space_id) 
	VALUES ('#fff9db', '#2f4c56', '5c12842c-61ce-4481-b33d-163d09a732c4');

`)

func _073_label_empty_name_sql() ([]byte, error) {
	return __073_label_empty_name_sql, nil
}

var __073_label_same_name_sql = []byte(`delete from labels;
INSERT INTO
	labels(name, text_color, background_color, space_id)
VALUES
   (
	'easy-fix', '#fff9db', '#2f4c56', '2e0698d8-753e-4cef-bb7c-f027634824a2'
   );
INSERT INTO
	labels(name, text_color, background_color, space_id)
VALUES
   (
	'easy-fix', '#fff9db', '#2f4c56', '2e0698d8-753e-4cef-bb7c-f027634824a2'
   );
`)

func _073_label_same_name_sql() ([]byte, error) {
	return __073_label_same_name_sql, nil
}

var __080_old_link_type_relics_sql = []byte(`SET spaces.system = '{{index . 0}}';

SET linktypes.bug_blocker = '{{index . 1}}';
SET linktypes.related = '{{index . 2}}';
SET linktypes.parenting = '{{index . 3}}';
-- we don't create this link type it just symbolizes a link type with no link type in existence
SET linktypes.completely_unknown = 'd30bb732-b277-48de-8d76-db241878bd30';

SET linktypes.unknown_bug_blocker = 'aad2a4ad-d601-4104-9804-2c977ca2e0c1';
SET linktypes.unknown_related = '355b647b-adc5-46b3-b297-cc54bc0554e6';
SET linktypes.unknown_parenting = '7479a9b9-8607-46fa-9535-d448fa8768ab';

SET cats.system = '{{index . 4}}';
SET cats.unknown_system = '75bc23dc-5aa3-4b1a-a3a6-b315e7ebeaa0';
SET cats.usercat = '{{index . 5}}';
SET cats.unknown_usercat = 'f83073d9-b79e-471b-a9a4-68248dd431ab';

INSERT INTO spaces (id, name) VALUES 
    (current_setting('spaces.system')::uuid, 'system.space')
    ON CONFLICT DO NOTHING;

INSERT INTO work_item_link_categories (id, name) VALUES
    (current_setting('cats.system')::uuid, 'system'),
    (current_setting('cats.unknown_system')::uuid, 'another system link category'),
    (current_setting('cats.usercat')::uuid, 'user'),
    (current_setting('cats.unknown_usercat')::uuid, 'another user link category')
    ON CONFLICT DO NOTHING;

INSERT INTO work_item_link_types (id, name, forward_name, reverse_name, topology, link_category_id, space_id) VALUES
    -- These are the known link types
    (   current_setting('linktypes.bug_blocker')::uuid,
        'Bug blocker', 'blocks', 'blocked by', 'network',
        current_setting('cats.system')::uuid, 
        current_setting('spaces.system')::uuid),

    (   current_setting('linktypes.related')::uuid, 
        'Related planner item', 'relates to', 'is related to', 'network', 
        current_setting('cats.system')::uuid, 
        current_setting('spaces.system')::uuid),

    (   current_setting('linktypes.parenting')::uuid,
        'Parent child item', 'parent of', 'child of', 'tree',
        current_setting('cats.system')::uuid,
        current_setting('spaces.system')::uuid),

    -- -- Insert the link types that exist in production but are left-overs from
    -- -- commit 90c595eaa02bde744207b6699d40ae4cc34a834e when I introduced fixed
    -- -- IDs for link types and categories. The following link types should be
    -- -- removed when we migrate to version 78 of the database.
    (   current_setting('linktypes.unknown_bug_blocker')::uuid,
        'Bug blocker', 'blocks', 'blocked by', 'network',
        current_setting('cats.unknown_system')::uuid, 
        current_setting('spaces.system')::uuid),

    (   current_setting('linktypes.unknown_related')::uuid, 
        'Related planner item', 'relates to', 'is related to', 'network', 
        current_setting('cats.unknown_system')::uuid, 
        current_setting('spaces.system')::uuid),

    (   current_setting('linktypes.unknown_parenting')::uuid,
        'Parent child item', 'parent of', 'child of', 'tree',
        current_setting('cats.unknown_system')::uuid,
        current_setting('spaces.system')::uuid);

-- Create some work items

SET wits.test = 'd998e454-08f7-48cb-97a0-c985073e77e2';
SET wis.parent1 = '95375720-4c50-4244-bd7c-04a7a33c4f28';
SET wis.parent2 = '4ab532d5-17fe-43e4-9c91-b62dddc3a02a';
SET wis.child1 = 'e7c3fab3-00a8-4ab8-9401-3545a92d5daa';
SET wis.child2 = '27d6c8e7-57a1-4dfa-8b05-3f2e3af9f5ca';

INSERT INTO work_item_types (id, name, space_id) VALUES
    (current_setting('wits.test')::uuid, 'Test WIT', current_setting('spaces.system')::uuid);

INSERT INTO work_items (id, space_id, type, fields) VALUES
    (current_setting('wis.parent1')::uuid, current_setting('spaces.system')::uuid, current_setting('wits.test')::uuid, '{"system.title":"Parent"}'::json),
    (current_setting('wis.parent2')::uuid, current_setting('spaces.system')::uuid, current_setting('wits.test')::uuid, '{"system.title":"Parent"}'::json),
    (current_setting('wis.child1')::uuid, current_setting('spaces.system')::uuid, current_setting('wits.test')::uuid, '{"system.title":"Child"}'::json),
    (current_setting('wis.child2')::uuid, current_setting('spaces.system')::uuid, current_setting('wits.test')::uuid, '{"system.title":"Child"}'::json);

-- Create links using both link types

SET ids.link1 = 'f87eb5dd-5021-4af1-9c60-79923dfa7ebd';
SET ids.link2 = '80f64aed-ecfe-45d7-b521-125d65ba4544';
SET ids.link3 = '77e25f85-77ff-4d53-a413-efc729a416e8';
SET ids.link4 = '45c02d6e-8196-4397-b1bd-ed354e978d4d';

INSERT INTO work_item_links (id, link_type_id, source_id, target_id) VALUES
    -- Create two links between the same WIs only using the old and new
    -- parenting link type. These links will be merged into one when migrating
    -- to version 78.
    (current_setting('ids.link1')::uuid, current_setting('linktypes.unknown_parenting')::uuid, current_setting('wis.parent1')::uuid, current_setting('wis.child1')::uuid),
    (current_setting('ids.link2')::uuid, current_setting('linktypes.parenting')::uuid, current_setting('wis.parent1')::uuid, current_setting('wis.child1')::uuid),
    -- This one will be changed to the new link type
    (current_setting('ids.link3')::uuid, current_setting('linktypes.unknown_parenting')::uuid, current_setting('wis.parent2')::uuid, current_setting('wis.child2')::uuid),
    -- This one exists because we need to create a link revision pointing to a
    -- valid link but using an unknown link type. 
    (current_setting('ids.link4')::uuid, current_setting('linktypes.related')::uuid, current_setting('wis.parent1')::uuid, current_setting('wis.child1')::uuid);

SET ids.user1 = 'e312b89c-0407-4fbb-b907-11d2ec37feec';
SET ids.user2 = 'ce5db17f-a047-41cf-b309-a64e9f293f4b';
SET ids.identity1 = '5f4f9360-c084-4039-8b94-7ea03e4d8fe1';
SET ids.identity2 = '42947554-229e-4f47-8880-ad0cad128da8';

SET modifierids.createid = '1';

-- users
INSERT INTO users(created_at, updated_at, id, email, full_name, image_url, bio, url, context_information)
VALUES (now(), now(), current_setting('ids.user1')::uuid, 'foobar1@example.com', 'test1', 'https://www.gravatar.com/avatar/testtwo2', 'my test bio one', 'http://example.com/001', '{"key": "value"}'),
(now(), now(), current_setting('ids.user2')::uuid, 'foobar2@example.com', 'test2', 'http://https://www.gravatar.com/avatar/testtwo3', 'my test bio two', 'http://example.com/002', '{"key": "value"}');

-- identities
INSERT INTO identities(created_at, updated_at, id, username, provider_type, user_id, profile_url)
VALUES (now(), now(), current_setting('ids.identity1')::uuid, 'test1', 'github', current_setting('ids.user1')::uuid, 'http://example-github.com/00123'),
(now(), now(), current_setting('ids.identity2')::uuid, 'test2', 'rhhd', current_setting('ids.user2')::uuid, 'http://example-rhd.com/00234');

-- Create appropriate link revisions for the create event
INSERT INTO work_item_link_revisions (
    revision_type, 
    modifier_id, 
    work_item_link_id, 
    work_item_link_version,
    work_item_link_type_id,
    work_item_link_source_id,
    work_item_link_target_id)
SELECT 
    current_setting('modifierids.createid')::int, 
    current_setting('ids.identity1')::uuid,
    id AS work_item_link_id,
    version AS work_item_link_version,
    link_type_id AS work_item_link_type_id,
    source_id AS work_item_link_source_id,
    target_id AS work_item_link_target_id
FROM work_item_links;

-- Manually update a link revision to point to a link type that doesn't exist.
UPDATE work_item_link_revisions SET work_item_link_type_id = current_setting('linktypes.completely_unknown')::uuid WHERE work_item_link_id = current_setting('ids.link4')::uuid;
`)

func _080_old_link_type_relics_sql() ([]byte, error) {
	return __080_old_link_type_relics_sql, nil
}

var __081_query_conflict_sql = []byte(`DELETE FROM spaces where id='e2297c54-eb0c-4eb0-b1f9-d3212cda5e1f';
INSERT INTO spaces (id, name) VALUES ('e2297c54-eb0c-4eb0-b1f9-d3212cda5e1f', 'test space');

DELETE FROM queries;
INSERT INTO
	queries(title, fields, space_id, creator)
VALUES
   (
	'assigned to me', '{"key": "value"}', 
    'e2297c54-eb0c-4eb0-b1f9-d3212cda5e1f',
    '5f6d7daf-6be3-4171-a77b-857b327c4bac'
   );
INSERT INTO
	queries(title, fields, space_id, creator)
VALUES
   (
	'assigned to me', '{"key": "value"}',
    'e2297c54-eb0c-4eb0-b1f9-d3212cda5e1f',
    '5f6d7daf-6be3-4171-a77b-857b327c4bac'
   );
`)

func _081_query_conflict_sql() ([]byte, error) {
	return __081_query_conflict_sql, nil
}

var __081_query_empty_title_sql = []byte(`DELETE FROM spaces where id='171d52ff-fa00-46d7-ac24-94269908ad7a';
INSERT INTO spaces (id, name) VALUES ('171d52ff-fa00-46d7-ac24-94269908ad7a', 'test space');

-- empty query name
-- this should fail
DELETE FROM queries;
INSERT INTO queries(title, fields, space_id, creator)
	VALUES ('', '{"assignee": "me"}', '171d52ff-fa00-46d7-ac24-94269908ad7a', '5ff348cf-57bc-4411-8812-21840107d25c');

`)

func _081_query_empty_title_sql() ([]byte, error) {
	return __081_query_empty_title_sql, nil
}

var __081_query_no_creator_sql = []byte(`DELETE FROM spaces where id='171d52ff-fa00-46d7-ac24-94269908ad7a';
INSERT INTO spaces (id, name) VALUES ('171d52ff-fa00-46d7-ac24-94269908ad7a', 'test space');

-- no creator ID provided
-- this should fail
DELETE FROM queries;
INSERT INTO queries(title, fields, space_id) 
	VALUES ('my queries', '{"assignee": "me"}', '171d52ff-fa00-46d7-ac24-94269908ad7a');

DELETE FROM spaces;
DELETE FROM queries;
`)

func _081_query_no_creator_sql() ([]byte, error) {
	return __081_query_no_creator_sql, nil
}

var __081_query_null_title_sql = []byte(`DELETE FROM spaces where id='171d52ff-fa00-46d7-ac24-94269908ad7a';
INSERT INTO spaces (id, name) VALUES ('171d52ff-fa00-46d7-ac24-94269908ad7a', 'test space');

-- empty query name
-- this should fail
DELETE FROM queries;
INSERT INTO queries(fields, space_id, creator) 
	VALUES ('{"assignee": "me"}', '171d52ff-fa00-46d7-ac24-94269908ad7a', '5ff348cf-57bc-4411-8812-21840107d25c');

`)

func _081_query_null_title_sql() ([]byte, error) {
	return __081_query_null_title_sql, nil
}

var __082_iteration_related_changes_sql = []byte(`insert into spaces (id, name) values ('11111111-8282-0000-0000-000000000000', 'test iteration - relationships changed at');
delete from work_items;
insert into work_item_types (id, name, space_id) values ('11111111-8282-0000-0000-000000000000', 'Test WIT','11111111-8282-0000-0000-000000000000');
insert into work_items (id, created_at, space_id, type, fields) values ('11111111-8282-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-8282-0000-0000-000000000000', '11111111-8282-0000-0000-000000000000', '{"system.title":"Work item 1"}'::json);
insert into work_items (id, created_at, space_id, type, fields) values ('22222222-8282-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-8282-0000-0000-000000000000', '11111111-8282-0000-0000-000000000000', '{"system.title":"Work item 2"}'::json);
insert into work_items (id, created_at, space_id, type, fields) values ('33333333-8282-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-8282-0000-0000-000000000000', '11111111-8282-0000-0000-000000000000', '{"system.title":"Work item 3"}'::json);
insert into work_items (id, created_at, space_id, type, fields) values ('44444444-8282-0000-0000-000000000000', (CURRENT_TIMESTAMP - interval '1 hour'), '11111111-8282-0000-0000-000000000000', '11111111-8282-0000-0000-000000000000', '{"system.title":"Work item 4"}'::json);

delete from iterations;
insert into iterations (id, name, created_at, space_id) values ('11111111-8282-0000-0000-000000000000', 'iteration 1', CURRENT_TIMESTAMP, '11111111-8282-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('22222222-8282-0000-0000-000000000000', 'iteration 2', CURRENT_TIMESTAMP, '11111111-8282-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('33333333-8282-0000-0000-000000000000', 'iteration 3', CURRENT_TIMESTAMP, '11111111-8282-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('44444444-8282-0000-0000-000000000000', 'iteration 4', CURRENT_TIMESTAMP, '11111111-8282-0000-0000-000000000000');
insert into iterations (id, name, created_at, space_id) values ('55555555-8282-0000-0000-000000000000', 'iteration 5', CURRENT_TIMESTAMP, '11111111-8282-0000-0000-000000000000');

-- link work item 1 to iteration 1
update work_items set updated_at = (CURRENT_TIMESTAMP + interval '1 hour'), fields = '{"system.title":"Work item 1", "system.iteration":"11111111-8282-0000-0000-000000000000"}'::json where id = '11111111-8282-0000-0000-000000000000';
-- link work item 2 to iteration 2 then iteration 3
update work_items set updated_at = (CURRENT_TIMESTAMP + interval '1 hour'), fields = '{"system.title":"Work item 2", "system.iteration":"22222222-8282-0000-0000-000000000000"}'::json where id = '22222222-8282-0000-0000-000000000000';
update work_items set updated_at = (CURRENT_TIMESTAMP + interval '2 hour'), fields = '{"system.title":"Work item 2", "system.iteration":"33333333-8282-0000-0000-000000000000"}'::json where id = '22222222-8282-0000-0000-000000000000';
-- link work item 3 to iteration 4 then soft-delete the work item
update work_items set fields = '{"system.title":"Work item 3", "system.iteration":"44444444-8282-0000-0000-000000000000"}'::json, updated_at = (CURRENT_TIMESTAMP + interval '1 hour') where id = '33333333-8282-0000-0000-000000000000';
update work_items set deleted_at = (CURRENT_TIMESTAMP + interval '2 hour') where id = '33333333-8282-0000-0000-000000000000';
-- link work item 4 to iteration 5 then set another, unrelated field
update work_items set fields = '{"system.title":"Work item 4", "system.iteration":"55555555-8282-0000-0000-000000000000"}'::json, updated_at = (CURRENT_TIMESTAMP + interval '1 hour') where id = '44444444-8282-0000-0000-000000000000';
update work_items set fields = '{"system.title":"Work item 4", "system.iteration":"55555555-8282-0000-0000-000000000000", "system.description":"foo"}'::json, updated_at = (CURRENT_TIMESTAMP + interval '2 hour') where id = '44444444-8282-0000-0000-000000000000';

`)

func _082_iteration_related_changes_sql() ([]byte, error) {
	return __082_iteration_related_changes_sql, nil
}

var __084_codebases_spaceid_url_idx_cleanup_sql = []byte(`-- Delete entries from codebases
DELETE FROM codebases
WHERE
    (url = 'https://github.com/fabric8-services/fabric8-wit/' AND
    space_id = '48d8987f-d7c2-454f-b67c-4cad74199b26');

-- Delete entries from spaces
DELETE FROM spaces
WHERE
    id = '48d8987f-d7c2-454f-b67c-4cad74199b26';
`)

func _084_codebases_spaceid_url_idx_cleanup_sql() ([]byte, error) {
	return __084_codebases_spaceid_url_idx_cleanup_sql, nil
}

var __084_codebases_spaceid_url_idx_setup_sql = []byte(`-- Create space for this test
INSERT INTO
    spaces(id, name)
VALUES
    ('48d8987f-d7c2-454f-b67c-4cad74199b26', 'foobarspace');


-- Insert in codebase an entry for spaceid and url that is deleted before
INSERT INTO
    codebases(url, space_id, id, deleted_at)
VALUES
    ('https://github.com/fabric8-services/fabric8-wit/',
     '48d8987f-d7c2-454f-b67c-4cad74199b26',
     '6bef707f-e2a4-4a39-95b8-8b51c4d8589d',
     '2018-03-10 05:28:33.262133+00');

-- Repeat the above sql command to create duplicate entry
-- but which is not deleted before.
INSERT INTO
    codebases(url, space_id, id)
VALUES
    ('https://github.com/fabric8-services/fabric8-wit/',
     '48d8987f-d7c2-454f-b67c-4cad74199b26',
     '97bd2bc3-106a-41d8-a1d9-3fdd6d2df1f7');

-- Insert in codebase an entry for spaceid and url that is deleted before
INSERT INTO
    codebases(url, space_id, id, deleted_at)
VALUES
    ('https://github.com/fabric8-services/fabric8-wit/',
     '48d8987f-d7c2-454f-b67c-4cad74199b26',
     '25feb91f-dbed-40ce-bb46-335a4d741213',
     '2018-03-10 05:28:33.262133+00');
`)

func _084_codebases_spaceid_url_idx_setup_sql() ([]byte, error) {
	return __084_codebases_spaceid_url_idx_setup_sql, nil
}

var __084_codebases_spaceid_url_idx_test_sql = []byte(`-- Try to create duplicate entry
SELECT *
FROM codebases
WHERE id='97bd2bc3-106a-41d8-a1d9-3fdd6d2df1f7';
`)

func _084_codebases_spaceid_url_idx_test_sql() ([]byte, error) {
	return __084_codebases_spaceid_url_idx_test_sql, nil
}

var __084_codebases_spaceid_url_idx_violate_sql = []byte(`-- Try to create duplicate entry
INSERT INTO
    codebases(url, space_id)
VALUES
    ('https://github.com/fabric8-services/fabric8-wit/',
     '48d8987f-d7c2-454f-b67c-4cad74199b26');
`)

func _084_codebases_spaceid_url_idx_violate_sql() ([]byte, error) {
	return __084_codebases_spaceid_url_idx_violate_sql, nil
}

var __085_delete_system_number_json_field_sql = []byte(`-- create space
insert into spaces (id, name) values ('c14f4210-e5e0-47c2-8be1-3482101009d9', 'test space');

-- create work item type 
INSERT INTO work_item_types (id,name,space_id,fields,description,icon) VALUES ('7fda71cb-c8be-4dfc-a872-f71573e48192', 'foo', 'c14f4210-e5e0-47c2-8be1-3482101009d9', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "system.number": {"Type": {"Kind": "integer"}, "Label": "Number", "Required": false, "Description": "The unique number that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create a work item
insert into work_items (id, type, space_id, fields) values ('27adc1a2-1ded-43b8-a125-12777139496c', '7fda71cb-c8be-4dfc-a872-f71573e48192', 'c14f4210-e5e0-47c2-8be1-3482101009d9', '{"system.title":"Work item 1", "system.number":1234}'::json);
insert into work_items (id, type, space_id, fields) values ('c106c056-2fec-4e56-83f0-cac31bb7ac1f', '7fda71cb-c8be-4dfc-a872-f71573e48192', 'c14f4210-e5e0-47c2-8be1-3482101009d9', '{"system.title":"Work item 2"}'::json);`)

func _085_delete_system_number_json_field_sql() ([]byte, error) {
	return __085_delete_system_number_json_field_sql, nil
}

var __093_codebases_add_cve_scan_cleanup_sql = []byte(`-- Delete the entries that were added for the space with id
DELETE FROM codebases
WHERE
    (space_id = 'ef8d5976-fe3a-4469-a575-8b704836450e');


-- Delete entries from spaces
DELETE FROM spaces
WHERE
    id = 'ef8d5976-fe3a-4469-a575-8b704836450e';
`)

func _093_codebases_add_cve_scan_cleanup_sql() ([]byte, error) {
	return __093_codebases_add_cve_scan_cleanup_sql, nil
}

var __093_codebases_add_cve_scan_setup_sql = []byte(`-- Create space for this test
INSERT INTO
    spaces(id, name, space_template_id)
VALUES
    ('ef8d5976-fe3a-4469-a575-8b704836450e', 'barfoospace', '929c963a-174c-4c37-b487-272067e88bd4');
`)

func _093_codebases_add_cve_scan_setup_sql() ([]byte, error) {
	return __093_codebases_add_cve_scan_setup_sql, nil
}

var __093_codebases_add_cve_scan_with_sql = []byte(`-- Insert in codebase an entry with cve_scan
INSERT INTO
    codebases(url, space_id, id, cve_scan)
VALUES
    ('https://github.com/fabric8-services/fabric8-auth/',
     'ef8d5976-fe3a-4469-a575-8b704836450e',
     '3a282f74-6bf3-4828-bcab-85ca3ff6a3ec',
     't');
`)

func _093_codebases_add_cve_scan_with_sql() ([]byte, error) {
	return __093_codebases_add_cve_scan_with_sql, nil
}

var __093_codebases_add_cve_scan_with2_sql = []byte(`-- Insert in codebase an entry with cve_scan
INSERT INTO
    codebases(url, space_id, id)
VALUES
    ('https://github.com/fabric8-services/fabric8-oso-proxy/',
     'ef8d5976-fe3a-4469-a575-8b704836450e',
     '1975cf7a-e3fb-4ca6-b368-bc8ee66fccee');
`)

func _093_codebases_add_cve_scan_with2_sql() ([]byte, error) {
	return __093_codebases_add_cve_scan_with2_sql, nil
}

var __093_codebases_add_cve_scan_without_sql = []byte(`-- Insert in codebase an entry without cve_scan
INSERT INTO
    codebases(url, space_id, id)
VALUES
    ('https://github.com/fabric8-services/fabric8-wit/',
     'ef8d5976-fe3a-4469-a575-8b704836450e',
     'cdc691c8-534d-48b3-9f72-7836bc5b9188');
`)

func _093_codebases_add_cve_scan_without_sql() ([]byte, error) {
	return __093_codebases_add_cve_scan_without_sql, nil
}

var __094_changes_to_agile_template_test_sql = []byte(`-- create space template
INSERT INTO space_templates (id,name,description) VALUES('f06ba0ba-eaf0-4655-bfe8-7b9e26d0f48f', 'test space template', 'test template');

-- create space
insert into spaces (id,name,space_template_id) values ('fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', 'test space', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f48f');

-- create work item type Theme
INSERT INTO work_item_types (id,name,space_template_id,fields,description,icon) VALUES ('5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a', 'Theme', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f48f', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "business_value": {"Type": {"Kind": "integer"}, "Label": "Business Value", "Required": false, "Description": "The business value of this work item."}, "effort": {"Type": {"Kind": "float"}, "Label": "Effort", "Required": false, "Description": "The effort that was given to this workitem within its space."}, "time_criticality": {"Type": {"Kind": "float"}, "Label": "Time Criticality", "Required": false, "Description": "The time criticality that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create work item type Epic
INSERT INTO work_item_types (id,name,space_template_id,fields,description,icon) VALUES ('2c169431-a55d-49eb-af74-cc19e895356f', 'Epic', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f48f', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "component": {"Type": {"Kind": "string"}, "Label": "Component", "Required": false, "Description": "The component value of this work item."}, "business_value": {"Type": {"Kind": "integer"}, "Label": "Business Value", "Required": false, "Description": "The business value of this work item."}, "effort": {"Type": {"Kind": "float"}, "Label": "Effort", "Required": false, "Description": "The effort that was given to this workitem within its space."}, "time_criticality": {"Type": {"Kind": "float"}, "Label": "Time Criticality", "Required": false, "Description": "The time criticality that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create work item type Story
INSERT INTO work_item_types (id,name,space_template_id,fields,description,icon) VALUES ('6ff83406-caa7-47a9-9200-4ca796be11bb', 'Story', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f48f', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "effort": {"Type": {"Kind": "float"}, "Label": "Effort", "Required": false, "Description": "The effort that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create a work items for Theme (removed fields 'effort' [float], 'business_value' [integer], 'time_criticality' [float])
insert into work_items (id, type, space_id, fields) values ('cf84c888-ac28-493d-a0cd-978b78568040', '5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', '{"system.title":"Work item 1", "effort":12.34, "business_value":1234, "time_criticality":56.78}'::json);
insert into work_items (id, type, space_id, fields) values ('8bbb542c-4f5c-44bb-9272-e1a8f24e6eb2', '5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', '{"system.title":"Work item 2"}'::json);

-- create a work items for Epic (removed fields 'effort' [float], 'business_value' [integer], 'time_criticality' [float], 'component' [string])
insert into work_items (id, type, space_id, fields) values ('4aebb314-a8c1-4e9c-96b6-074769d16934', '2c169431-a55d-49eb-af74-cc19e895356f', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', '{"system.title":"Work item 3", "effort":12.34, "business_value":1234, "time_criticality":56.78, "component":"Component 1"}'::json);
insert into work_items (id, type, space_id, fields) values ('9c53fb2b-c6af-48a1-bef1-6fa547ea72fa', '2c169431-a55d-49eb-af74-cc19e895356f', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', '{"system.title":"Work item 4"}'::json);

-- create a work items for Story (removed fields 'effort' [float])
insert into work_items (id, type, space_id, fields) values ('68f83154-8d76-49c1-8be0-063ce90f803d', '6ff83406-caa7-47a9-9200-4ca796be11bb', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', '{"system.title":"Work item 5", "effort":12.34}'::json);
insert into work_items (id, type, space_id, fields) values ('17e2081f-812d-4f4e-9c51-c537406bd1d8', '6ff83406-caa7-47a9-9200-4ca796be11bb', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b4', '{"system.title":"Work item 6"}'::json);
`)

func _094_changes_to_agile_template_test_sql() ([]byte, error) {
	return __094_changes_to_agile_template_test_sql, nil
}

var __095_remove_resolution_field_from_impediment_sql = []byte(`-- create space template
INSERT INTO space_templates (id,name,description) VALUES('2b542bba-e131-412a-8762-88688e1423f2', 'test space template 2b542bba-e131-412a-8762-88688e1423f2', 'test template');

-- create space
insert into spaces (id,name,space_template_id) values ('b12cf363-3625-4b7b-99f9-159c497202e2', 'test space b12cf363-3625-4b7b-99f9-159c497202e2', '2b542bba-e131-412a-8762-88688e1423f2');

-- create work item type Impediment
INSERT INTO work_item_types (id,name,can_construct,space_template_id,fields,description,icon) VALUES ('03b9bb64-4f65-4fa7-b165-494cd4f01401', 'Impediment', 'false', '2b542bba-e131-412a-8762-88688e1423f2', '{"resolution": {"type": {"values": ["Done", "Rejected", "Duplicate", "Incomplete Description", "Can not Reproduce", "Partially Completed", "Deferred", "Wont Fix", "Out of Date", "Explained", "Verified"], "base_type": {"kind": "string"}, "simple_type": {"kind": "enum"}, "rewritable_values": false}, "label": "Resolution", "required": false, "read_only": false, "description": "The reason why this work items state was last changed.\n"}, "system.area": {"type": {"kind": "area"}, "label": "Area", "required": false, "read_only": false, "description": "The area to which the work item belongs"}, "system.order": {"type": {"kind": "float"}, "label": "Execution Order", "required": false, "read_only": true, "description": "Execution Order of the workitem"}, "system.state": {"type": {"values": ["New", "Open", "In Progress", "Resolved", "Closed"], "base_type": {"kind": "string"}, "simple_type": {"kind": "enum"}, "rewritable_values": false}, "label": "State", "required": true, "read_only": false, "description": "The state of the impediment."}, "system.title": {"type": {"kind": "string"}, "label": "Title", "required": true, "read_only": false, "description": "The title text of the work item"}, "system.labels": {"type": {"simple_type": {"kind": "list"}, "component_type": {"kind": "label"}}, "label": "Labels", "required": false, "read_only": false, "description": "List of labels attached to the work item"}, "system.number": {"type": {"kind": "integer"}, "label": "Number", "required": false, "read_only": true, "description": "The unique number that was given to this workitem within its space."}, "system.creator": {"type": {"kind": "user"}, "label": "Creator", "required": true, "read_only": false, "description": "The user that created the work item"}, "system.codebase": {"type": {"kind": "codebase"}, "label": "Codebase", "required": false, "read_only": false, "description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"type": {"simple_type": {"kind": "list"}, "component_type": {"kind": "user"}}, "label": "Assignees", "required": false, "read_only": false, "description": "The users that are assigned to the work item"}, "system.iteration": {"type": {"kind": "iteration"}, "label": "Iteration", "required": false, "read_only": false, "description": "The iteration to which the work item belongs"}, "system.created_at": {"type": {"kind": "instant"}, "label": "Created at", "required": false, "read_only": true, "description": "The date and time when the work item was created"}, "system.updated_at": {"type": {"kind": "instant"}, "label": "Updated at", "required": false, "read_only": true, "description": "The date and time when the work item was last updated"}, "system.description": {"type": {"kind": "markup"}, "label": "Description", "required": false, "read_only": false, "description": "A descriptive text of the work item"}, "system.remote_item_id": {"type": {"kind": "string"}, "label": "Remote item", "required": false, "read_only": false, "description": "The ID of the remote work item"}}', 'Description for Impediment', 'fa fa-bookmark');

-- Create a few work items for Impediment - one with and one without a
-- resolution set.
insert into work_items (id, type, space_id, fields) values ('24ed462d-0430-4ffe-ba4f-7b5725b6a48c', '03b9bb64-4f65-4fa7-b165-494cd4f01401', 'b12cf363-3625-4b7b-99f9-159c497202e2', '{"system.title":"Work item 1", "resolution":"Rejected"}'::json); -- this resolution does only exist in the former version of the agile template

insert into work_items (id, type, space_id, fields) values ('6a870ee3-e57c-4f98-9c7a-3cdf2ef5c2ef', '03b9bb64-4f65-4fa7-b165-494cd4f01401', 'b12cf363-3625-4b7b-99f9-159c497202e2', '{"system.title":"Work item 2"}'::json); -- no resolution specified`)

func _095_remove_resolution_field_from_impediment_sql() ([]byte, error) {
	return __095_remove_resolution_field_from_impediment_sql, nil
}

var __096_changes_to_agile_template_test_sql = []byte(`-- create space template
INSERT INTO space_templates (id,name,description) VALUES('f06ba0ba-eaf0-4655-bfe8-7b9e26d0f444', 'test space template f06ba0ba-eaf0-4655-bfe8-7b9e26d0f444', 'test template');

-- create space
insert into spaces (id,name,space_template_id) values ('fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', 'test space fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f444');

-- create work item type Theme
DELETE FROM work_item_types WHERE id = '5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a';
INSERT INTO work_item_types (id,name,space_template_id,fields,description,icon) VALUES ('5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a', 'Theme', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f444', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "business_value": {"Type": {"Kind": "integer"}, "Label": "Business Value", "Required": false, "Description": "The business value of this work item."}, "effort": {"Type": {"Kind": "float"}, "Label": "Effort", "Required": false, "Description": "The effort that was given to this workitem within its space."}, "time_criticality": {"Type": {"Kind": "float"}, "Label": "Time Criticality", "Required": false, "Description": "The time criticality that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create work item type Epic
DELETE FROM work_item_types WHERE id = '2c169431-a55d-49eb-af74-cc19e895356f';
INSERT INTO work_item_types (id,name,space_template_id,fields,description,icon) VALUES ('2c169431-a55d-49eb-af74-cc19e895356f', 'Epic', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f444', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "component": {"Type": {"Kind": "string"}, "Label": "Component", "Required": false, "Description": "The component value of this work item."}, "business_value": {"Type": {"Kind": "integer"}, "Label": "Business Value", "Required": false, "Description": "The business value of this work item."}, "effort": {"Type": {"Kind": "float"}, "Label": "Effort", "Required": false, "Description": "The effort that was given to this workitem within its space."}, "time_criticality": {"Type": {"Kind": "float"}, "Label": "Time Criticality", "Required": false, "Description": "The time criticality that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create work item type Story
DELETE FROM work_item_types WHERE id = '6ff83406-caa7-47a9-9200-4ca796be11bb';
INSERT INTO work_item_types (id,name,space_template_id,fields,description,icon) VALUES ('6ff83406-caa7-47a9-9200-4ca796be11bb', 'Story', 'f06ba0ba-eaf0-4655-bfe8-7b9e26d0f444', '{"system.area": {"Type": {"Kind": "area"}, "Label": "Area", "Required": false, "Description": "The area to which the work item belongs"}, "system.order": {"Type": {"Kind": "float"}, "Label": "Execution Order", "Required": false, "Description": "Execution Order of the workitem."}, "system.state": {"Type": {"Kind": "enum", "Values": ["new", "open", "in progress", "resolved", "closed"], "BaseType": {"Kind": "string"}}, "Label": "State", "Required": true, "Description": "The state of the work item"}, "system.title": {"Type": {"Kind": "string"}, "Label": "Title", "Required": true, "Description": "The title text of the work item"}, "system.labels": {"Type": {"Kind": "list", "ComponentType": {"Kind": "label"}}, "Label": "Labels", "Required": false, "Description": "List of labels attached to the work item"}, "effort": {"Type": {"Kind": "float"}, "Label": "Effort", "Required": false, "Description": "The effort that was given to this workitem within its space."}, "system.creator": {"Type": {"Kind": "user"}, "Label": "Creator", "Required": true, "Description": "The user that created the work item"}, "system.codebase": {"Type": {"Kind": "codebase"}, "Label": "Codebase", "Required": false, "Description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"Type": {"Kind": "list", "ComponentType": {"Kind": "user"}}, "Label": "Assignees", "Required": false, "Description": "The users that are assigned to the work item"}, "system.iteration": {"Type": {"Kind": "iteration"}, "Label": "Iteration", "Required": false, "Description": "The iteration to which the work item belongs"}, "system.created_at": {"Type": {"Kind": "instant"}, "Label": "Created at", "Required": false, "Description": "The date and time when the work item was created"}, "system.updated_at": {"Type": {"Kind": "instant"}, "Label": "Updated at", "Required": false, "Description": "The date and time when the work item was last updated"}, "system.description": {"Type": {"Kind": "markup"}, "Label": "Description", "Required": false, "Description": "A descriptive text of the work item"}, "system.remote_item_id": {"Type": {"Kind": "string"}, "Label": "Remote item", "Required": false, "Description": "The ID of the remote work item"}}', 'Description for Planner Item', 'fa fa-bookmark');

-- create a work items for Theme (removed fields 'effort' [float], 'business_value' [integer], 'time_criticality' [float])
-- create space template
insert into work_items (id, type, space_id, fields) values ('cf84c888-ac28-493d-a0cd-978b78568011', '5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', '{"system.title":"Work item 1", "effort":12.34, "business_value":1234, "time_criticality":56.78}'::json);
insert into work_items (id, type, space_id, fields) values ('8bbb542c-4f5c-44bb-9272-e1a8f24e6e22', '5182fc8c-b1d6-4c3d-83ca-6a3c781fa18a', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', '{"system.title":"Work item 2"}'::json);

-- create a work items for Epic (removed fields 'effort' [float], 'business_value' [integer], 'time_criticality' [float], 'component' [string])
insert into work_items (id, type, space_id, fields) values ('4aebb314-a8c1-4e9c-96b6-074769d16933', '2c169431-a55d-49eb-af74-cc19e895356f', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', '{"system.title":"Work item 3", "effort":12.34, "business_value":1234, "time_criticality":56.78, "component":"Component 1"}'::json);
insert into work_items (id, type, space_id, fields) values ('9c53fb2b-c6af-48a1-bef1-6fa547ea7244', '2c169431-a55d-49eb-af74-cc19e895356f', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', '{"system.title":"Work item 4"}'::json);

-- create a work items for Story (removed fields 'effort' [float])
insert into work_items (id, type, space_id, fields) values ('68f83154-8d76-49c1-8be0-063ce90f8055', '6ff83406-caa7-47a9-9200-4ca796be11bb', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', '{"system.title":"Work item 5", "effort":12.34}'::json);
insert into work_items (id, type, space_id, fields) values ('17e2081f-812d-4f4e-9c51-c537406bd166', '6ff83406-caa7-47a9-9200-4ca796be11bb', 'fe8e7e07-a8d7-41c2-9761-ca1ffe2409b5', '{"system.title":"Work item 6"}'::json);
`)

func _096_changes_to_agile_template_test_sql() ([]byte, error) {
	return __096_changes_to_agile_template_test_sql, nil
}

var __097_remove_resolution_field_from_impediment_sql = []byte(`-- create space template
INSERT INTO space_templates (id,name,description) VALUES('2b542bba-e131-412a-8762-88688e142311', 'test space template 2b542bba-e131-412a-8762-88688e142311', 'test template');

-- create space
insert into spaces (id,name,space_template_id) values ('b12cf363-3625-4b7b-99f9-159c49720200', 'test space b12cf363-3625-4b7b-99f9-159c49720200', '2b542bba-e131-412a-8762-88688e142311');

-- create work item type Impediment
DELETE FROM work_item_types WHERE id = '03b9bb64-4f65-4fa7-b165-494cd4f01401';
INSERT INTO work_item_types (id,name,can_construct,space_template_id,fields,description,icon) VALUES ('03b9bb64-4f65-4fa7-b165-494cd4f01401', 'Impediment', 'false', '2b542bba-e131-412a-8762-88688e142311', '{"resolution": {"type": {"values": ["Done", "Rejected", "Duplicate", "Incomplete Description", "Can not Reproduce", "Partially Completed", "Deferred", "Wont Fix", "Out of Date", "Explained", "Verified"], "base_type": {"kind": "string"}, "simple_type": {"kind": "enum"}, "rewritable_values": false}, "label": "Resolution", "required": false, "read_only": false, "description": "The reason why this work items state was last changed.\n"}, "system.area": {"type": {"kind": "area"}, "label": "Area", "required": false, "read_only": false, "description": "The area to which the work item belongs"}, "system.order": {"type": {"kind": "float"}, "label": "Execution Order", "required": false, "read_only": true, "description": "Execution Order of the workitem"}, "system.state": {"type": {"values": ["New", "Open", "In Progress", "Resolved", "Closed"], "base_type": {"kind": "string"}, "simple_type": {"kind": "enum"}, "rewritable_values": false}, "label": "State", "required": true, "read_only": false, "description": "The state of the impediment."}, "system.title": {"type": {"kind": "string"}, "label": "Title", "required": true, "read_only": false, "description": "The title text of the work item"}, "system.labels": {"type": {"simple_type": {"kind": "list"}, "component_type": {"kind": "label"}}, "label": "Labels", "required": false, "read_only": false, "description": "List of labels attached to the work item"}, "system.number": {"type": {"kind": "integer"}, "label": "Number", "required": false, "read_only": true, "description": "The unique number that was given to this workitem within its space."}, "system.creator": {"type": {"kind": "user"}, "label": "Creator", "required": true, "read_only": false, "description": "The user that created the work item"}, "system.codebase": {"type": {"kind": "codebase"}, "label": "Codebase", "required": false, "read_only": false, "description": "Contains codebase attributes to which this WI belongs to"}, "system.assignees": {"type": {"simple_type": {"kind": "list"}, "component_type": {"kind": "user"}}, "label": "Assignees", "required": false, "read_only": false, "description": "The users that are assigned to the work item"}, "system.iteration": {"type": {"kind": "iteration"}, "label": "Iteration", "required": false, "read_only": false, "description": "The iteration to which the work item belongs"}, "system.created_at": {"type": {"kind": "instant"}, "label": "Created at", "required": false, "read_only": true, "description": "The date and time when the work item was created"}, "system.updated_at": {"type": {"kind": "instant"}, "label": "Updated at", "required": false, "read_only": true, "description": "The date and time when the work item was last updated"}, "system.description": {"type": {"kind": "markup"}, "label": "Description", "required": false, "read_only": false, "description": "A descriptive text of the work item"}, "system.remote_item_id": {"type": {"kind": "string"}, "label": "Remote item", "required": false, "read_only": false, "description": "The ID of the remote work item"}}', 'Description for Impediment', 'fa fa-bookmark');

-- Create a few work items for Impediment - one with and one without a
-- resolution set.
insert into work_items (id, type, space_id, fields) values ('24ed462d-0430-4ffe-ba4f-7b5725b6a411', '03b9bb64-4f65-4fa7-b165-494cd4f01401', 'b12cf363-3625-4b7b-99f9-159c49720200', '{"system.title":"Work item 1", "resolution":"Rejected"}'::json); -- this resolution does only exist in the former version of the agile template

insert into work_items (id, type, space_id, fields) values ('6a870ee3-e57c-4f98-9c7a-3cdf2ef5c222', '03b9bb64-4f65-4fa7-b165-494cd4f01401', 'b12cf363-3625-4b7b-99f9-159c49720200', '{"system.title":"Work item 2"}'::json); -- no resolution specified`)

func _097_remove_resolution_field_from_impediment_sql() ([]byte, error) {
	return __097_remove_resolution_field_from_impediment_sql, nil
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
	"044-insert-test-data.sql": _044_insert_test_data_sql,
	"045-update-work-items.sql": _045_update_work_items_sql,
	"046-insert-oauth-states.sql": _046_insert_oauth_states_sql,
	"047-insert-codebases.sql": _047_insert_codebases_sql,
	"048-unique-idx-failed-insert.sql": _048_unique_idx_failed_insert_sql,
	"050-users-add-column-company.sql": _050_users_add_column_company_sql,
	"053-edit-username.sql": _053_edit_username_sql,
	"054-add-stackid-to-codebase.sql": _054_add_stackid_to_codebase_sql,
	"055-assign-root-area-if-missing.sql": _055_assign_root_area_if_missing_sql,
	"056-assign-root-iteration-if-missing.sql": _056_assign_root_iteration_if_missing_sql,
	"057-add-last-used-workspace-to-codebase.sql": _057_add_last_used_workspace_to_codebase_sql,
	"061-add-duplicate-space-owner-name.sql": _061_add_duplicate_space_owner_name_sql,
	"063-workitem-related-changes.sql": _063_workitem_related_changes_sql,
	"065-workitem-id-unique-per-space.sql": _065_workitem_id_unique_per_space_sql,
	"066-work_item_links_data_integrity.sql": _066_work_item_links_data_integrity_sql,
	"067-comment-parentid-uuid.sql": _067_comment_parentid_uuid_sql,
	"071-iteration-related-changes.sql": _071_iteration_related_changes_sql,
	"073-label-color-code.sql": _073_label_color_code_sql,
	"073-label-color-code2.sql": _073_label_color_code2_sql,
	"073-label-empty-name.sql": _073_label_empty_name_sql,
	"073-label-same-name.sql": _073_label_same_name_sql,
	"080-old-link-type-relics.sql": _080_old_link_type_relics_sql,
	"081-query-conflict.sql": _081_query_conflict_sql,
	"081-query-empty-title.sql": _081_query_empty_title_sql,
	"081-query-no-creator.sql": _081_query_no_creator_sql,
	"081-query-null-title.sql": _081_query_null_title_sql,
	"082-iteration-related-changes.sql": _082_iteration_related_changes_sql,
	"084-codebases-spaceid-url-idx-cleanup.sql": _084_codebases_spaceid_url_idx_cleanup_sql,
	"084-codebases-spaceid-url-idx-setup.sql": _084_codebases_spaceid_url_idx_setup_sql,
	"084-codebases-spaceid-url-idx-test.sql": _084_codebases_spaceid_url_idx_test_sql,
	"084-codebases-spaceid-url-idx-violate.sql": _084_codebases_spaceid_url_idx_violate_sql,
	"085-delete-system.number-json-field.sql": _085_delete_system_number_json_field_sql,
	"093-codebases-add-cve-scan-cleanup.sql": _093_codebases_add_cve_scan_cleanup_sql,
	"093-codebases-add-cve-scan-setup.sql": _093_codebases_add_cve_scan_setup_sql,
	"093-codebases-add-cve-scan-with.sql": _093_codebases_add_cve_scan_with_sql,
	"093-codebases-add-cve-scan-with2.sql": _093_codebases_add_cve_scan_with2_sql,
	"093-codebases-add-cve-scan-without.sql": _093_codebases_add_cve_scan_without_sql,
	"094-changes-to-agile-template-test.sql": _094_changes_to_agile_template_test_sql,
	"095-remove-resolution-field-from-impediment.sql": _095_remove_resolution_field_from_impediment_sql,
	"096-changes-to-agile-template-test.sql": _096_changes_to_agile_template_test_sql,
	"097-remove-resolution-field-from-impediment.sql": _097_remove_resolution_field_from_impediment_sql,
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
	"044-insert-test-data.sql": &_bintree_t{_044_insert_test_data_sql, map[string]*_bintree_t{
	}},
	"045-update-work-items.sql": &_bintree_t{_045_update_work_items_sql, map[string]*_bintree_t{
	}},
	"046-insert-oauth-states.sql": &_bintree_t{_046_insert_oauth_states_sql, map[string]*_bintree_t{
	}},
	"047-insert-codebases.sql": &_bintree_t{_047_insert_codebases_sql, map[string]*_bintree_t{
	}},
	"048-unique-idx-failed-insert.sql": &_bintree_t{_048_unique_idx_failed_insert_sql, map[string]*_bintree_t{
	}},
	"050-users-add-column-company.sql": &_bintree_t{_050_users_add_column_company_sql, map[string]*_bintree_t{
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
	"061-add-duplicate-space-owner-name.sql": &_bintree_t{_061_add_duplicate_space_owner_name_sql, map[string]*_bintree_t{
	}},
	"063-workitem-related-changes.sql": &_bintree_t{_063_workitem_related_changes_sql, map[string]*_bintree_t{
	}},
	"065-workitem-id-unique-per-space.sql": &_bintree_t{_065_workitem_id_unique_per_space_sql, map[string]*_bintree_t{
	}},
	"066-work_item_links_data_integrity.sql": &_bintree_t{_066_work_item_links_data_integrity_sql, map[string]*_bintree_t{
	}},
	"067-comment-parentid-uuid.sql": &_bintree_t{_067_comment_parentid_uuid_sql, map[string]*_bintree_t{
	}},
	"071-iteration-related-changes.sql": &_bintree_t{_071_iteration_related_changes_sql, map[string]*_bintree_t{
	}},
	"073-label-color-code.sql": &_bintree_t{_073_label_color_code_sql, map[string]*_bintree_t{
	}},
	"073-label-color-code2.sql": &_bintree_t{_073_label_color_code2_sql, map[string]*_bintree_t{
	}},
	"073-label-empty-name.sql": &_bintree_t{_073_label_empty_name_sql, map[string]*_bintree_t{
	}},
	"073-label-same-name.sql": &_bintree_t{_073_label_same_name_sql, map[string]*_bintree_t{
	}},
	"080-old-link-type-relics.sql": &_bintree_t{_080_old_link_type_relics_sql, map[string]*_bintree_t{
	}},
	"081-query-conflict.sql": &_bintree_t{_081_query_conflict_sql, map[string]*_bintree_t{
	}},
	"081-query-empty-title.sql": &_bintree_t{_081_query_empty_title_sql, map[string]*_bintree_t{
	}},
	"081-query-no-creator.sql": &_bintree_t{_081_query_no_creator_sql, map[string]*_bintree_t{
	}},
	"081-query-null-title.sql": &_bintree_t{_081_query_null_title_sql, map[string]*_bintree_t{
	}},
	"082-iteration-related-changes.sql": &_bintree_t{_082_iteration_related_changes_sql, map[string]*_bintree_t{
	}},
	"084-codebases-spaceid-url-idx-cleanup.sql": &_bintree_t{_084_codebases_spaceid_url_idx_cleanup_sql, map[string]*_bintree_t{
	}},
	"084-codebases-spaceid-url-idx-setup.sql": &_bintree_t{_084_codebases_spaceid_url_idx_setup_sql, map[string]*_bintree_t{
	}},
	"084-codebases-spaceid-url-idx-test.sql": &_bintree_t{_084_codebases_spaceid_url_idx_test_sql, map[string]*_bintree_t{
	}},
	"084-codebases-spaceid-url-idx-violate.sql": &_bintree_t{_084_codebases_spaceid_url_idx_violate_sql, map[string]*_bintree_t{
	}},
	"085-delete-system.number-json-field.sql": &_bintree_t{_085_delete_system_number_json_field_sql, map[string]*_bintree_t{
	}},
	"093-codebases-add-cve-scan-cleanup.sql": &_bintree_t{_093_codebases_add_cve_scan_cleanup_sql, map[string]*_bintree_t{
	}},
	"093-codebases-add-cve-scan-setup.sql": &_bintree_t{_093_codebases_add_cve_scan_setup_sql, map[string]*_bintree_t{
	}},
	"093-codebases-add-cve-scan-with.sql": &_bintree_t{_093_codebases_add_cve_scan_with_sql, map[string]*_bintree_t{
	}},
	"093-codebases-add-cve-scan-with2.sql": &_bintree_t{_093_codebases_add_cve_scan_with2_sql, map[string]*_bintree_t{
	}},
	"093-codebases-add-cve-scan-without.sql": &_bintree_t{_093_codebases_add_cve_scan_without_sql, map[string]*_bintree_t{
	}},
	"094-changes-to-agile-template-test.sql": &_bintree_t{_094_changes_to_agile_template_test_sql, map[string]*_bintree_t{
	}},
	"095-remove-resolution-field-from-impediment.sql": &_bintree_t{_095_remove_resolution_field_from_impediment_sql, map[string]*_bintree_t{
	}},
	"096-changes-to-agile-template-test.sql": &_bintree_t{_096_changes_to_agile_template_test_sql, map[string]*_bintree_t{
	}},
	"097-remove-resolution-field-from-impediment.sql": &_bintree_t{_097_remove_resolution_field_from_impediment_sql, map[string]*_bintree_t{
	}},
}}
