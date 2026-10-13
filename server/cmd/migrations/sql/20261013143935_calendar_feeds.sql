-- Modify "absences" table
ALTER TABLE "public"."absences" ADD COLUMN "source" text NOT NULL DEFAULT 'manual', ADD COLUMN "external_uid" text NULL;
-- Create index "idx_absences_person_external_uid" to table: "absences"
CREATE UNIQUE INDEX "idx_absences_person_external_uid" ON "public"."absences" ("person_id", "external_uid") WHERE ((external_uid IS NOT NULL) AND (deleted_at IS NULL));
-- Create "calendar_feeds" table
CREATE TABLE "public"."calendar_feeds" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "person_id" uuid NOT NULL,
  "url_ciphertext" bytea NOT NULL,
  "host" text NOT NULL,
  "last_synced_at" timestamptz NULL,
  "last_error" text NULL,
  "events_found" bigint NOT NULL DEFAULT 0,
  "next_sync_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_calendar_feeds_person" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_calendar_feeds_deleted_at" to table: "calendar_feeds"
CREATE INDEX "idx_calendar_feeds_deleted_at" ON "public"."calendar_feeds" ("deleted_at");
-- Create index "idx_calendar_feeds_next_sync_at" to table: "calendar_feeds"
CREATE INDEX "idx_calendar_feeds_next_sync_at" ON "public"."calendar_feeds" ("next_sync_at");
-- Create index "idx_calendar_feeds_person_id" to table: "calendar_feeds"
CREATE UNIQUE INDEX "idx_calendar_feeds_person_id" ON "public"."calendar_feeds" ("person_id");
