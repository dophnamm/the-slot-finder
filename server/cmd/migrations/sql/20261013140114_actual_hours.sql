-- Create "actual_hours" table
CREATE TABLE "public"."actual_hours" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "person_id" uuid NOT NULL,
  "project_id" uuid NOT NULL,
  "week_start" date NOT NULL,
  "hours" numeric(6,2) NOT NULL,
  "updated_by" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_actual_hours_person" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_actual_hours_project" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_actual_hours_deleted_at" to table: "actual_hours"
CREATE INDEX "idx_actual_hours_deleted_at" ON "public"."actual_hours" ("deleted_at");
-- Create index "idx_actual_hours_person_project_week" to table: "actual_hours"
CREATE UNIQUE INDEX "idx_actual_hours_person_project_week" ON "public"."actual_hours" ("person_id", "project_id", "week_start");
-- Create index "idx_actual_hours_project_id" to table: "actual_hours"
CREATE INDEX "idx_actual_hours_project_id" ON "public"."actual_hours" ("project_id");
-- Create index "idx_actual_hours_week" to table: "actual_hours"
CREATE INDEX "idx_actual_hours_week" ON "public"."actual_hours" ("week_start");
