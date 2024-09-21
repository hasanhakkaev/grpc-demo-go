
BEGIN;

ALTER TABLE yqapp_demo_schema.tasks DROP COLUMN state;
DROP TYPE state_enum;

COMMIT;