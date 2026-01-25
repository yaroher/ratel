CREATE TABLE currency
(
    code char(3) PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE users
(
    user_id    bigserial PRIMARY KEY,
    email      text        NOT NULL UNIQUE,
    full_name  text        NOT NULL,
    is_active  boolean     NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE products
(
    product_id bigserial PRIMARY KEY,
    sku        text           NOT NULL UNIQUE,
    name       text           NOT NULL,
    price      NUMERIC(12, 2) NOT NULL,
    currency   char(3)        NOT NULL REFERENCES currency (code),
    stock_qty  integer        NOT NULL DEFAULT 0 CHECK (stock_qty >= 0),
    is_deleted boolean        NOT NULL DEFAULT false, -- soft delete
    created_at timestamptz    NOT NULL DEFAULT now(),
    updated_at timestamptz    NOT NULL DEFAULT now()
);

CREATE INDEX ix_products_not_deleted ON products (product_id) WHERE is_deleted = false;


CREATE TABLE orders
(
    order_id   bigserial PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES users (user_id),
    status     text        NOT NULL CHECK (status IN ('NEW', 'PAID', 'CANCELLED', 'SHIPPED')),
    currency   char(3)     NOT NULL REFERENCES currency (code),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_orders_user_created ON orders (user_id, created_at DESC);

CREATE TABLE order_items
(
    order_id   bigint         NOT NULL REFERENCES orders (order_id) ON DELETE CASCADE,
    line_no    integer        NOT NULL CHECK (line_no > 0),
    product_id bigint         NOT NULL REFERENCES products (product_id),
    qty        integer        NOT NULL CHECK (qty > 0),
    unit_price NUMERIC(12, 2) NOT NULL,
    PRIMARY KEY (order_id, line_no),
    -- защита от дублей одного товара в одном заказе (по желанию)
    UNIQUE (order_id, product_id)
);

CREATE INDEX ix_order_items_product ON order_items (product_id);
