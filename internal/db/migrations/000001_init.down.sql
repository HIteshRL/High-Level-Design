-- 000001_init.down.sql
DROP TRIGGER IF EXISTS set_updated_at_conversations ON conversations;
DROP TRIGGER IF EXISTS set_updated_at_users ON users;
DROP FUNCTION IF EXISTS trigger_set_updated_at();
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS message_role;
DROP TYPE IF EXISTS user_role;
