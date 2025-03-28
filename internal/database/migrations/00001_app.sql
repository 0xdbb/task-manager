-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE "user" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" VARCHAR NOT NULL CHECK (LENGTH(name) > 0),
  "email" VARCHAR UNIQUE NOT NULL,
  "password" VARCHAR NOT NULL,
  "role" VARCHAR(20) NOT NULL,
  "created_at" TIMESTAMPTZ DEFAULT now(),
  "updated_at" TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE "task" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" UUID NOT NULL,
  "type" VARCHAR(50) NOT NULL,
  "payload" TEXT NOT NULL,
  "status" VARCHAR(20) DEFAULT 'pending',
  "result" TEXT,
  "created_at" TIMESTAMP DEFAULT now(),
  "updated_at" TIMESTAMP,
  FOREIGN KEY ("user_id") REFERENCES "user" ("id")
);

CREATE TABLE "task_log" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "task_id" UUID NOT NULL,
  "worker_id" VARCHAR(100),
  "status" VARCHAR(20) NOT NULL,
  "message" TEXT,
  "created_at" TIMESTAMP DEFAULT now(),
  FOREIGN KEY ("task_id") REFERENCES "task" ("id")
);

CREATE TABLE "notification" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" UUID,
  "task_id" UUID,
  "message" VARCHAR(255),
  "read" BOOLEAN DEFAULT false,
  "created_at" TIMESTAMP DEFAULT now(),
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

-- Drop extension
DROP EXTENSION IF EXISTS pgcrypto;

-- +goose StatementEnd

