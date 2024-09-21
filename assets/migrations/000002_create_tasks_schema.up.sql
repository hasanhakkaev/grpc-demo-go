CREATE SCHEMA IF NOT EXISTS yqapp_demo_schema;

-- Grant usage and creation rights to the yqapp_demo_schema schema
GRANT USAGE ON SCHEMA yqapp_demo_schema TO yqapp_demo_user;

-- Allow yqapp_demo_user to create tables, sequences, etc., in the yqapp_demo_schema schema
GRANT CREATE ON SCHEMA yqapp_demo_schema TO yqapp_demo_user;

-- Grant all privileges on all tables and sequences (for future tables) in the yqapp_demo_schema schema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA yqapp_demo_schema TO yqapp_demo_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA yqapp_demo_schema TO yqapp_demo_user;

-- Ensure future tables and sequences created in the yqapp_demo_schema schema automatically grant privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA yqapp_demo_schema
    GRANT ALL PRIVILEGES ON TABLES TO yqapp_demo_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA yqapp_demo_schema
    GRANT ALL PRIVILEGES ON SEQUENCES TO yqapp_demo_user;
