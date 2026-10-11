-- Modify "plans" table
ALTER TABLE "public"."plans" ADD COLUMN "superseded_by" uuid NULL;
-- Create "change_events" table
CREATE TABLE "public"."change_events" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "type" text NOT NULL,
  "payload" jsonb NOT NULL DEFAULT '{}',
  "occurred_at" timestamptz NOT NULL,
  "status" text NOT NULL,
  "available_at" timestamptz NOT NULL,
  "attempts" bigint NOT NULL DEFAULT 0,
  "last_error" text NULL,
  "draft_id" uuid NULL,
  "project_ids" jsonb NOT NULL DEFAULT '[]',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_change_events_draft" FOREIGN KEY ("draft_id") REFERENCES "public"."plans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_change_events_deleted_at" to table: "change_events"
CREATE INDEX "idx_change_events_deleted_at" ON "public"."change_events" ("deleted_at");
-- Create index "idx_change_events_draft_id" to table: "change_events"
CREATE INDEX "idx_change_events_draft_id" ON "public"."change_events" ("draft_id");
-- Create index "idx_change_events_status_available" to table: "change_events"
CREATE INDEX "idx_change_events_status_available" ON "public"."change_events" ("status", "available_at");
-- Create index "idx_change_events_type" to table: "change_events"
CREATE INDEX "idx_change_events_type" ON "public"."change_events" ("type");
