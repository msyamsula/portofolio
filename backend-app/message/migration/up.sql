-- Migration: event-sourced message system schema
--
-- Design principles:
--   1. event_log is append-only and never mutated — it is the source of truth.
--   2. State is always derived by replaying events in version order.
--   3. event_log doubles as the outbox (published=false means pending Kafka publish).
--   4. conversation is the terminal "done" marker; its presence means delivery succeeded.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─────────────────────────────────────────────────────────────────────────────
-- event_log: every state change produces exactly one row here.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS event_log (
    event_id        UUID        NOT NULL,
    message_id      UUID        NOT NULL,
    conversation_id UUID        NOT NULL,
    sender_id       UUID        NOT NULL,
    receiver_id     UUID        NOT NULL,

    -- version is the optimistic concurrency counter for this (message_id, conversation_id).
    -- Two concurrent processes reading the same max(version) and racing to insert
    -- the same (message_id, conversation_id, version) triple: only ONE wins.
    -- The loser receives 0 rows affected and treats it as an idempotent no-op.
    version         INT         NOT NULL,

    -- event_name is the directed edge label in the state machine.
    -- Valid: SENT | SUCCESS | FAILED | RECONCILE_STUCKED_MESSAGE | RECONCILE_FAILED_MESSAGE
    event_name      TEXT        NOT NULL,

    -- payload carries the original message body. Copied from SENT into all
    -- subsequent events so that every event is self-describing.
    payload         JSONB,

    -- published: false until the outbox relay has forwarded this event to Kafka.
    published       BOOLEAN     NOT NULL DEFAULT FALSE,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT event_log_pkey          PRIMARY KEY (event_id),
    CONSTRAINT event_log_version_uniq  UNIQUE      (message_id, conversation_id, version)
);

-- Outbox relay: efficiently find events not yet sent to Kafka.
CREATE INDEX IF NOT EXISTS idx_event_log_outbox
    ON event_log (created_at ASC)
    WHERE published = FALSE;

-- State reconstruction: ordered scan per (message_id, conversation_id).
CREATE INDEX IF NOT EXISTS idx_event_log_message
    ON event_log (message_id, conversation_id, version ASC);

-- ─────────────────────────────────────────────────────────────────────────────
-- conversation: written atomically with the SUCCESS event.
-- A row here is the authoritative terminal "done" state.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS conversation (
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL,
    sender_id       UUID        NOT NULL,
    receiver_id     UUID        NOT NULL,
    message_id      UUID        NOT NULL,

    -- event_id ties the row to the exact SUCCESS event that created it.
    -- Its uniqueness prevents duplicate conversation rows from racing SUCCESS inserts.
    event_id        UUID        NOT NULL,

    payload         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT conversation_pkey         PRIMARY KEY (id),
    CONSTRAINT conversation_event_uniq   UNIQUE      (event_id),
    CONSTRAINT conversation_message_uniq UNIQUE      (message_id, conversation_id)
);
