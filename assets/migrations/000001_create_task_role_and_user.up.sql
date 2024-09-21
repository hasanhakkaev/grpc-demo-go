CREATE ROLE yqapp_demo_role;
CREATE user yqapp_demo_user WITH LOGIN PASSWORD 'myS3cr3tP4ssw0rd';
GRANT yqapp_demo_role TO yqapp_demo_user;
