-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE "user" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "first_name" varchar NOT NULL CHECK (LENGTH(first_name) > 0),
  "last_name" varchar NOT NULL CHECK (LENGTH(last_name) > 0),
  "email" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "phone" varchar NOT NULL,
  "address" varchar NOT NULL,
  "date_of_birth" varchar NOT NULL,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "store" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" varchar UNIQUE NOT NULL,
  "address" varchar NOT NULL,
  "city" varchar NOT NULL,
  "description" text NOT NULL,
  "admin_id" uuid NOT NULL,
  "created_at" timestamptz DEFAULT (now()), 
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "category" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" varchar UNIQUE NOT NULL,
  "created_at" timestamptz DEFAULT (now()), 
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "product" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" varchar NOT NULL,
  "description" text NOT NULL,
  "image_url" varchar NOT NULL,
  "category_id" uuid NOT NULL,
  "price" decimal(10,2) NOT NULL,
  "stock" int NOT NULL,
  "store_id" uuid NOT NULL,
  "created_at" timestamptz DEFAULT (now()), 
  "updated_at" timestamptz DEFAULT (now()),
  UNIQUE (name, store_id)  -- Ensures product names are unique within a store
);


CREATE TABLE "order" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "status" varchar NOT NULL,
  "delivery_date" date,
  "created_at" timestamp DEFAULT (now()), 
  "updated_at" timestamp DEFAULT (now())
);

CREATE TABLE "order_item" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "quantity" int NOT NULL,
  "total_price" decimal(10,2) NOT NULL,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now()) 
);

CREATE TABLE "review" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "rating" int,
  "comment" text,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "payment" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "amount" decimal(10,2),
  "payment_method" varchar,
  "payment_date" timestamptz,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "notification" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "message" text,
  "sent_at" timestamptz,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now()) 
);

CREATE TABLE "social_media" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "platform" varchar,
  "link" varchar,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "testimonial" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "message" text,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

-- Indexes for frequently searched columns
CREATE INDEX ON "user" ("email");
CREATE INDEX ON "user" ("address");
CREATE INDEX ON "store" ("name");
CREATE INDEX ON "store" ("address");
CREATE INDEX ON "category" ("name");
CREATE INDEX ON "product" ("name");
CREATE INDEX ON "order_item" ("quantity");
CREATE INDEX ON "order_item" ("total_price");
CREATE INDEX ON "review" ("rating");
CREATE INDEX ON "review" ("comment");
CREATE INDEX ON "payment" ("amount");
CREATE INDEX ON "payment" ("payment_method");
CREATE INDEX ON "notification" ("message");
CREATE INDEX ON "social_media" ("platform");
CREATE INDEX ON "social_media" ("link");
CREATE INDEX ON "testimonial" ("message");

-- ✅ Foreign Keys with Indexes
ALTER TABLE "store" ADD CONSTRAINT store_admin_id_fkey FOREIGN KEY ("admin_id") REFERENCES "user" ("id");
CREATE INDEX idx_store_admin_id ON "store" ("admin_id");

ALTER TABLE "product" ADD CONSTRAINT product_category_id_fkey FOREIGN KEY ("category_id") REFERENCES "category" ("id");
CREATE INDEX idx_product_category_id ON "product" ("category_id");

ALTER TABLE "product" ADD CONSTRAINT product_store_id_fkey FOREIGN KEY ("store_id") REFERENCES "store" ("id");
CREATE INDEX idx_product_store_id ON "product" ("store_id");

ALTER TABLE "order" ADD CONSTRAINT order_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "user" ("id");
CREATE INDEX idx_order_user_id ON "order" ("user_id");

ALTER TABLE "order_item" ADD CONSTRAINT order_item_order_id_fkey FOREIGN KEY ("order_id") REFERENCES "order" ("id");
CREATE INDEX idx_order_item_order_id ON "order_item" ("order_id");

ALTER TABLE "order_item" ADD CONSTRAINT order_item_product_id_fkey FOREIGN KEY ("product_id") REFERENCES "product" ("id");
CREATE INDEX idx_order_item_product_id ON "order_item" ("product_id");

ALTER TABLE "review" ADD CONSTRAINT review_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "user" ("id");
CREATE INDEX idx_review_user_id ON "review" ("user_id");

ALTER TABLE "review" ADD CONSTRAINT review_product_id_fkey FOREIGN KEY ("product_id") REFERENCES "product" ("id");
CREATE INDEX idx_review_product_id ON "review" ("product_id");

ALTER TABLE "payment" ADD CONSTRAINT payment_order_id_fkey FOREIGN KEY ("order_id") REFERENCES "order" ("id");
CREATE INDEX idx_payment_order_id ON "payment" ("order_id");

ALTER TABLE "notification" ADD CONSTRAINT notification_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "user" ("id");
CREATE INDEX idx_notification_user_id ON "notification" ("user_id");

ALTER TABLE "social_media" ADD CONSTRAINT social_media_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "user" ("id");
CREATE INDEX idx_social_media_user_id ON "social_media" ("user_id");

ALTER TABLE "testimonial" ADD CONSTRAINT testimonial_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "user" ("id");
CREATE INDEX idx_testimonial_user_id ON "testimonial" ("user_id");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop Foreign Keys First
ALTER TABLE "store" DROP CONSTRAINT store_admin_id_fkey;
ALTER TABLE "product" DROP CONSTRAINT product_category_id_fkey;
ALTER TABLE "product" DROP CONSTRAINT product_store_id_fkey;
ALTER TABLE "order" DROP CONSTRAINT order_user_id_fkey;
ALTER TABLE "order_item" DROP CONSTRAINT order_item_order_id_fkey;
ALTER TABLE "order_item" DROP CONSTRAINT order_item_product_id_fkey;
ALTER TABLE "review" DROP CONSTRAINT review_user_id_fkey;
ALTER TABLE "review" DROP CONSTRAINT review_product_id_fkey;
ALTER TABLE "payment" DROP CONSTRAINT payment_order_id_fkey;
ALTER TABLE "notification" DROP CONSTRAINT notification_user_id_fkey;
ALTER TABLE "social_media" DROP CONSTRAINT social_media_user_id_fkey;
ALTER TABLE "testimonial" DROP CONSTRAINT testimonial_user_id_fkey;

-- Drop Indexes
DROP INDEX IF EXISTS idx_store_admin_id;
DROP INDEX IF EXISTS idx_product_category_id;
DROP INDEX IF EXISTS idx_product_store_id;
DROP INDEX IF EXISTS idx_order_user_id;
DROP INDEX IF EXISTS idx_order_item_order_id;
DROP INDEX IF EXISTS idx_order_item_product_id;
DROP INDEX IF EXISTS idx_review_user_id;
DROP INDEX IF EXISTS idx_review_product_id;
DROP INDEX IF EXISTS idx_payment_order_id;
DROP INDEX IF EXISTS idx_notification_user_id;
DROP INDEX IF EXISTS idx_social_media_user_id;
DROP INDEX IF EXISTS idx_testimonial_user_id;

-- Drop Tables in Reverse Order
DROP TABLE IF EXISTS "testimonial";
DROP TABLE IF EXISTS "social_media";
DROP TABLE IF EXISTS "notification";
DROP TABLE IF EXISTS "payment";
DROP TABLE IF EXISTS "review";
DROP TABLE IF EXISTS "order_item";
DROP TABLE IF EXISTS "order";
DROP TABLE IF EXISTS "product";
DROP TABLE IF EXISTS "category";
DROP TABLE IF EXISTS "store";
DROP TABLE IF EXISTS "user";

-- Drop Extension
DROP EXTENSION IF EXISTS pgcrypto;
-- +goose StatementEnd

