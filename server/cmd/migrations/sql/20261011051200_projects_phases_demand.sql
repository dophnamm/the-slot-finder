-- Create "projects" table
CREATE TABLE "public"."projects" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "program_id" uuid NOT NULL,
  "name" text NOT NULL,
  "priority" bigint NOT NULL DEFAULT 3,
  "status" text NOT NULL,
  "start_earliest" date NOT NULL,
  "deadline" date NOT NULL,
  "scope_points" numeric(8,2) NOT NULL DEFAULT 0,
  "done_points" numeric(8,2) NOT NULL DEFAULT 0,
  "lead_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_projects_lead" FOREIGN KEY ("lead_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "fk_projects_program" FOREIGN KEY ("program_id") REFERENCES "public"."programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_projects_deleted_at" to table: "projects"
CREATE INDEX "idx_projects_deleted_at" ON "public"."projects" ("deleted_at");
-- Create index "idx_projects_lead_id" to table: "projects"
CREATE INDEX "idx_projects_lead_id" ON "public"."projects" ("lead_id");
-- Create index "idx_projects_program_id" to table: "projects"
CREATE INDEX "idx_projects_program_id" ON "public"."projects" ("program_id");
-- Create index "idx_projects_status" to table: "projects"
CREATE INDEX "idx_projects_status" ON "public"."projects" ("status");
-- Create "project_phases" table
CREATE TABLE "public"."project_phases" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "project_id" uuid NOT NULL,
  "name" text NOT NULL,
  "sequence" bigint NOT NULL,
  "start_week_start" date NOT NULL,
  "end_week_start" date NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_projects_phases" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_project_phases_deleted_at" to table: "project_phases"
CREATE INDEX "idx_project_phases_deleted_at" ON "public"."project_phases" ("deleted_at");
-- Create index "idx_project_phases_order" to table: "project_phases"
CREATE INDEX "idx_project_phases_order" ON "public"."project_phases" ("project_id", "sequence");
-- Create "demands" table
CREATE TABLE "public"."demands" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "project_id" uuid NOT NULL,
  "expertise_id" uuid NOT NULL,
  "phase_id" uuid NULL,
  "min_level" bigint NOT NULL,
  "headcount" bigint NOT NULL DEFAULT 1,
  "total_hours" numeric(8,2) NOT NULL,
  "profile" text NOT NULL DEFAULT 'flat',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_demands_expertise" FOREIGN KEY ("expertise_id") REFERENCES "public"."expertises" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "fk_demands_phase" FOREIGN KEY ("phase_id") REFERENCES "public"."project_phases" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "fk_projects_demands" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_demands_deleted_at" to table: "demands"
CREATE INDEX "idx_demands_deleted_at" ON "public"."demands" ("deleted_at");
-- Create index "idx_demands_phase_id" to table: "demands"
CREATE INDEX "idx_demands_phase_id" ON "public"."demands" ("phase_id");
-- Create index "idx_demands_project_id" to table: "demands"
CREATE INDEX "idx_demands_project_id" ON "public"."demands" ("project_id");
