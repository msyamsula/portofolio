-- Rollback: drop in reverse creation order.
DROP INDEX  IF EXISTS idx_event_log_message;
DROP INDEX  IF EXISTS idx_event_log_outbox;
DROP TABLE  IF EXISTS conversation;
DROP TABLE  IF EXISTS event_log;
