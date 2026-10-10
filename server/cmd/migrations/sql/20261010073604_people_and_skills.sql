-- Create "people" table
CREATE TABLE "public"."people" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "email" text NOT NULL,
  "keycloak_id" text NULL,
  "role" text NOT NULL,
  "weekly_hours" numeric(5,2) NOT NULL,
  "non_project_hours" numeric(5,2) NOT NULL DEFAULT 0,
  "start_date" date NULL,
  "end_date" date NULL,
  "active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_people_deleted_at" to table: "people"
CREATE INDEX "idx_people_deleted_at" ON "public"."people" ("deleted_at");
-- Create index "idx_people_email" to table: "people"
CREATE UNIQUE INDEX "idx_people_email" ON "public"."people" ("email");
-- Create index "idx_people_keycloak_id" to table: "people"
CREATE UNIQUE INDEX "idx_people_keycloak_id" ON "public"."people" ("keycloak_id");
-- Create "expertises" table
CREATE TABLE "public"."expertises" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "sequence" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "idx_expertises_deleted_at" to table: "expertises"
CREATE INDEX "idx_expertises_deleted_at" ON "public"."expertises" ("deleted_at");
-- Create index "idx_expertises_name" to table: "expertises"
CREATE UNIQUE INDEX "idx_expertises_name" ON "public"."expertises" ("name");
-- Create "person_skills" table
CREATE TABLE "public"."person_skills" (
  "person_id" uuid NOT NULL,
  "expertise_id" uuid NOT NULL,
  "level" bigint NOT NULL,
  "is_primary" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("person_id", "expertise_id"),
  CONSTRAINT "fk_people_skills" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_person_skills_expertise" FOREIGN KEY ("expertise_id") REFERENCES "public"."expertises" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "idx_person_skills_one_primary" to table: "person_skills"
CREATE UNIQUE INDEX "idx_person_skills_one_primary" ON "public"."person_skills" ("person_id") WHERE is_primary;
