-- Create "programs" table
CREATE TABLE "public"."programs" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "client" text NOT NULL,
  "status" text NOT NULL,
  "start_date" date NOT NULL,
  "end_date" date NOT NULL,
  "goal" text NOT NULL DEFAULT '',
  "objectives" jsonb NOT NULL DEFAULT '[]',
  "charter_url" text NULL,
  "lead_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_programs_lead" FOREIGN KEY ("lead_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "idx_programs_deleted_at" to table: "programs"
CREATE INDEX "idx_programs_deleted_at" ON "public"."programs" ("deleted_at");
-- Create index "idx_programs_lead_id" to table: "programs"
CREATE INDEX "idx_programs_lead_id" ON "public"."programs" ("lead_id");
-- Create index "idx_programs_status" to table: "programs"
CREATE INDEX "idx_programs_status" ON "public"."programs" ("status");
