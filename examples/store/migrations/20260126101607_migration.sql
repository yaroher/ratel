-- Create "currency" table
CREATE TABLE "public"."currency" ("code" character(3) NOT NULL, "name" text NOT NULL, PRIMARY KEY ("code"));
-- Create "users" table
CREATE TABLE "public"."users" ("user_id" bigserial NOT NULL, "email" text NOT NULL, "full_name" text NOT NULL, "is_active" boolean NOT NULL DEFAULT true, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("user_id"), CONSTRAINT "users_email_key" UNIQUE ("email"));
-- Create "orders" table
CREATE TABLE "public"."orders" ("order_id" bigserial NOT NULL, "user_id" bigint NOT NULL, "status" text NOT NULL, "currency" character(3) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("order_id"), CONSTRAINT "orders_currency_fkey" FOREIGN KEY ("currency") REFERENCES "public"."currency" ("code") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "orders_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("user_id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "orders_status_check" CHECK (status = ANY (ARRAY['NEW'::text, 'PAID'::text, 'CANCELLED'::text, 'SHIPPED'::text])));
-- Create index "ix_orders_user_created" to table: "orders"
CREATE INDEX "ix_orders_user_created" ON "public"."orders" ("user_id", "created_at" DESC);
-- Create "products" table
CREATE TABLE "public"."products" ("product_id" bigserial NOT NULL, "sku" text NOT NULL, "name" text NOT NULL, "price" numeric(12,2) NOT NULL, "currency" character(3) NOT NULL, "stock_qty" integer NOT NULL DEFAULT 0, "is_deleted" boolean NOT NULL DEFAULT false, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("product_id"), CONSTRAINT "products_sku_key" UNIQUE ("sku"), CONSTRAINT "products_currency_fkey" FOREIGN KEY ("currency") REFERENCES "public"."currency" ("code") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "products_stock_qty_check" CHECK (stock_qty >= 0));
-- Create index "ix_products_not_deleted" to table: "products"
CREATE INDEX "ix_products_not_deleted" ON "public"."products" ("product_id") WHERE (is_deleted = false);
-- Create "order_items" table
CREATE TABLE "public"."order_items" ("order_id" bigint NOT NULL, "line_no" integer NOT NULL, "product_id" bigint NOT NULL, "qty" integer NOT NULL, "unit_price" numeric(12,2) NOT NULL, CONSTRAINT "order_items_order_id_product_id_key" UNIQUE ("order_id", "product_id"), CONSTRAINT "order_items_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "public"."orders" ("order_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "order_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("product_id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "order_items_line_no_check" CHECK (line_no > 0), CONSTRAINT "order_items_qty_check" CHECK (qty > 0));
-- Create index "ix_order_items_product" to table: "order_items"
CREATE INDEX "ix_order_items_product" ON "public"."order_items" ("product_id");
-- Create "categories" table
CREATE TABLE "public"."categories" ("category_id" bigserial NOT NULL, "name" text NOT NULL, "slug" text NOT NULL, "parent_id" bigint NULL, PRIMARY KEY ("category_id"), CONSTRAINT "categories_slug_key" UNIQUE ("slug"), CONSTRAINT "categories_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."categories" ("category_id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create "product_categories" table
CREATE TABLE "public"."product_categories" ("product_id" bigint NOT NULL, "category_id" bigint NOT NULL, PRIMARY KEY ("product_id", "category_id"), CONSTRAINT "product_categories_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("category_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "product_categories_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("product_id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "tags" table
CREATE TABLE "public"."tags" ("tag_id" bigserial NOT NULL, "name" text NOT NULL, "slug" text NOT NULL, PRIMARY KEY ("tag_id"), CONSTRAINT "tags_slug_key" UNIQUE ("slug"));
-- Create "product_tags" table
CREATE TABLE "public"."product_tags" ("product_id" bigint NOT NULL, "tag_id" bigint NOT NULL, PRIMARY KEY ("product_id", "tag_id"), CONSTRAINT "product_tags_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products" ("product_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "product_tags_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."tags" ("tag_id") ON UPDATE NO ACTION ON DELETE CASCADE);
