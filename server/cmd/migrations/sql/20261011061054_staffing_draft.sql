-- Modify "projects" table
ALTER TABLE "public"."projects" ADD COLUMN "staffed_at" timestamptz NULL;
-- Create "audit_events" table
CREATE TABLE "public"."audit_events" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "actor_id" uuid NULL,
  "actor_subject" text NOT NULL,
  "action" text NOT NULL,
  "resource" text NOT NULL,
  "resource_id" uuid NOT NULL,
  "before" jsonb NULL,
  "after" jsonb NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_audit_events_action" to table: "audit_events"
CREATE INDEX "idx_audit_events_action" ON "public"."audit_events" ("action");
-- Create index "idx_audit_events_actor_id" to table: "audit_events"
CREATE INDEX "idx_audit_events_actor_id" ON "public"."audit_events" ("actor_id");
-- Create index "idx_audit_events_deleted_at" to table: "audit_events"
CREATE INDEX "idx_audit_events_deleted_at" ON "public"."audit_events" ("deleted_at");
-- Create index "idx_audit_events_resource_id" to table: "audit_events"
CREATE INDEX "idx_audit_events_resource_id" ON "public"."audit_events" ("resource_id");
-- Create "plans" table
CREATE TABLE "public"."plans" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "parent_id" uuid NULL,
  "kind" text NOT NULL,
  "status" text NOT NULL,
  "version" bigint NOT NULL DEFAULT 0,
  "project_ids" jsonb NOT NULL DEFAULT '[]',
  "unmet" jsonb NOT NULL DEFAULT '[]',
  "created_by" uuid NULL,
  "published_at" timestamptz NULL,
  "published_by" uuid NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_plans_deleted_at" to table: "plans"
CREATE INDEX "idx_plans_deleted_at" ON "public"."plans" ("deleted_at");
-- Create index "idx_plans_parent_id" to table: "plans"
CREATE INDEX "idx_plans_parent_id" ON "public"."plans" ("parent_id");
-- Create index "idx_plans_status" to table: "plans"
CREATE INDEX "idx_plans_status" ON "public"."plans" ("status");
-- Create index "plan_one_published_baseline" to table: "plans"
CREATE UNIQUE INDEX "plan_one_published_baseline" ON "public"."plans" ("kind") WHERE ((kind = 'baseline'::text) AND (status = 'published'::text));
-- Create "assignments" table
CREATE TABLE "public"."assignments" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "plan_id" uuid NOT NULL,
  "person_id" uuid NOT NULL,
  "demand_id" uuid NOT NULL,
  "kind" text NOT NULL,
  "reason" jsonb NOT NULL DEFAULT '[]',
  "created_by" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_assignments_demand" FOREIGN KEY ("demand_id") REFERENCES "public"."demands" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_assignments_person" FOREIGN KEY ("person_id") REFERENCES "public"."people" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "fk_plans_assignments" FOREIGN KEY ("plan_id") REFERENCES "public"."plans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_assignments_deleted_at" to table: "assignments"
CREATE INDEX "idx_assignments_deleted_at" ON "public"."assignments" ("deleted_at");
-- Create index "idx_assignments_demand_id" to table: "assignments"
CREATE INDEX "idx_assignments_demand_id" ON "public"."assignments" ("demand_id");
-- Create index "idx_assignments_person_id" to table: "assignments"
CREATE INDEX "idx_assignments_person_id" ON "public"."assignments" ("person_id");
-- Create index "idx_assignments_plan_person_demand" to table: "assignments"
CREATE UNIQUE INDEX "idx_assignments_plan_person_demand" ON "public"."assignments" ("plan_id", "person_id", "demand_id");
-- Create "allocations" table
CREATE TABLE "public"."allocations" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "assignment_id" uuid NOT NULL,
  "week_start" date NOT NULL,
  "hours" numeric(6,2) NOT NULL,
  "locked" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_assignments_allocations" FOREIGN KEY ("assignment_id") REFERENCES "public"."assignments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_allocations_assignment_week" to table: "allocations"
CREATE UNIQUE INDEX "idx_allocations_assignment_week" ON "public"."allocations" ("assignment_id", "week_start");
-- Create index "idx_allocations_deleted_at" to table: "allocations"
CREATE INDEX "idx_allocations_deleted_at" ON "public"."allocations" ("deleted_at");
-- Create index "idx_allocations_week" to table: "allocations"
CREATE INDEX "idx_allocations_week" ON "public"."allocations" ("week_start");
