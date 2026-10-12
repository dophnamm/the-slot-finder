-- Modify "plans" table
ALTER TABLE "public"."plans" ADD COLUMN "name" text NULL, ADD COLUMN "assumptions" jsonb NOT NULL DEFAULT '[]', ADD COLUMN "summary" jsonb NULL;
