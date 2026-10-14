-- Create "program_team_members" table
CREATE TABLE "public"."program_team_members" (
  "program_id" uuid NOT NULL,
  "person_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("program_id", "person_id"),
  CONSTRAINT "fk_program_team_members_person" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_program_team_members_program" FOREIGN KEY ("program_id") REFERENCES "public"."programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_program_team_members_person_id" to table: "program_team_members"
CREATE INDEX "idx_program_team_members_person_id" ON "public"."program_team_members" ("person_id");
