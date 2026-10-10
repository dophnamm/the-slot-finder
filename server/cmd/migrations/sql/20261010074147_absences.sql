-- Create "absences" table
CREATE TABLE "public"."absences" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "person_id" uuid NULL,
  "from_date" date NOT NULL,
  "to_date" date NOT NULL,
  "hours_per_day" numeric(4,2) NOT NULL,
  "type" text NOT NULL,
  "status" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_absences_person" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_absences_dates" to table: "absences"
CREATE INDEX "idx_absences_dates" ON "public"."absences" ("from_date", "to_date");
-- Create index "idx_absences_deleted_at" to table: "absences"
CREATE INDEX "idx_absences_deleted_at" ON "public"."absences" ("deleted_at");
-- Create index "idx_absences_person_id" to table: "absences"
CREATE INDEX "idx_absences_person_id" ON "public"."absences" ("person_id");
