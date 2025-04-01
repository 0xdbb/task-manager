-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create ENUM types
CREATE TYPE user_role AS ENUM ('ADMIN', 'STANDARD', 'WORKER');
CREATE TYPE task_status AS ENUM ('PENDING', 'IN-PROGRESS', 'FAILED', 'COMPLETED');
CREATE TYPE task_priority AS ENUM ('LOW', 'MEDIUM', 'HIGH');

CREATE TABLE "user" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" VARCHAR NOT NULL CHECK (LENGTH(name) > 0),
  "email" VARCHAR UNIQUE NOT NULL,
  "password" VARCHAR NOT NULL,
  "role" user_role NOT NULL,
  "created_at" TIMESTAMPTZ DEFAULT now(),
  "updated_at" TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE "task" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "title" VARCHAR NOT NULL,
  "type" VARCHAR(50) NOT NULL,
  "description" VARCHAR NOT NULL,
  "user_id" UUID NOT NULL,
  "priority" task_priority NOT NULL,
  "payload" TEXT NOT NULL,
  "status" task_status DEFAULT 'PENDING',
  "result" TEXT,
  "due_time" TIMESTAMPTZ NOT NULL, 
  "created_at" TIMESTAMPTZ DEFAULT now(),
  "updated_at" TIMESTAMPTZ DEFAULT now(),
  FOREIGN KEY ("user_id") REFERENCES "user" ("id")
);

CREATE TABLE "task_log" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "task_id" UUID NOT NULL,
  "worker_id" VARCHAR(100),
  "status" task_status NOT NULL,
  "message" TEXT,
  "created_at" TIMESTAMPTZ DEFAULT now(),
  FOREIGN KEY ("task_id") REFERENCES "task" ("id")
);

CREATE TABLE "notification" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" UUID,
  "task_id" UUID,
  "message" VARCHAR(255),
  "sent" BOOLEAN DEFAULT false,
  "created_at" TIMESTAMPTZ DEFAULT now(),
  FOREIGN KEY ("user_id") REFERENCES "user" ("id"),
  FOREIGN KEY ("task_id") REFERENCES "task" ("id")
);

-- Indexes
CREATE INDEX ON "user" ("email");
CREATE INDEX ON "task" ("user_id");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop foreign key constraints explicitly
ALTER TABLE "notification" DROP CONSTRAINT IF EXISTS "notification_user_id_fkey";
ALTER TABLE "notification" DROP CONSTRAINT IF EXISTS "notification_task_id_fkey";
ALTER TABLE "task_log" DROP CONSTRAINT IF EXISTS "task_log_task_id_fkey";
ALTER TABLE "task" DROP CONSTRAINT IF EXISTS "task_user_id_fkey";

-- Drop tables in reverse order
DROP TABLE IF EXISTS "notification";
DROP TABLE IF EXISTS "task_log";
DROP TABLE IF EXISTS "task";
DROP TABLE IF EXISTS "user";

-- Drop ENUM types
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS task_priority;

-- Drop extension
DROP EXTENSION IF EXISTS pgcrypto;

-- +goose StatementEnd

