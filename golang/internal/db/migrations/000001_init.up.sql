CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    title VARCHAR(255),
    role VARCHAR(50) NOT NULL,
    owner_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE branches (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    manager_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE dishes (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    description TEXT,
    image VARCHAR(255),
    status VARCHAR(50) DEFAULT 'Available',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sets (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    user_id BIGINT,
    is_favourite BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    image VARCHAR(255),
    price INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE set_dishes (
    set_id BIGINT REFERENCES sets(id) ON DELETE CASCADE,
    dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (set_id, dish_id)
);

CREATE TYPE table_status AS ENUM ('AVAILABLE', 'OCCUPIED', 'RESERVED', 'OUT_OF_SERVICE', 'TAKE_AWAY');

CREATE TABLE tables (
    number INTEGER PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    capacity INTEGER NOT NULL,
    status table_status DEFAULT 'AVAILABLE',
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE guests (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    table_number INTEGER REFERENCES tables(number) ON DELETE SET NULL,
    refresh_token VARCHAR(255),
    refresh_token_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    guest_id BIGINT REFERENCES guests(id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    is_guest BOOLEAN DEFAULT FALSE,
    table_number INTEGER REFERENCES tables(number) ON DELETE SET NULL,
    order_handler_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'Pending',
    total_price INTEGER,
    topping VARCHAR(255),
    tracking_order VARCHAR(255),
    take_away BOOLEAN DEFAULT FALSE,
    chili_number INTEGER,
    table_token VARCHAR(255),
    order_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_dishes (
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (order_id, dish_id)
);

CREATE TABLE order_sets (
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    set_id BIGINT REFERENCES sets(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (order_id, set_id)
);

CREATE TABLE deliveries (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    guest_id BIGINT REFERENCES guests(id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    is_guest BOOLEAN DEFAULT FALSE,
    table_number INTEGER REFERENCES tables(number) ON DELETE SET NULL,
    order_handler_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    status VARCHAR(50),
    total_price INTEGER,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    bow_chili INTEGER,
    bow_no_chili INTEGER,
    take_away BOOLEAN DEFAULT FALSE,
    chili_number INTEGER,
    table_token VARCHAR(255),
    client_name VARCHAR(255),
    delivery_address VARCHAR(255),
    delivery_contact VARCHAR(255),
    delivery_notes TEXT,
    scheduled_time TIMESTAMP WITH TIME ZONE,
    delivery_fee INTEGER,
    delivery_status VARCHAR(50),
    estimated_delivery_time TIMESTAMP WITH TIME ZONE,
    actual_delivery_time TIMESTAMP WITH TIME ZONE,  
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE delivery_dishes (
    delivery_id BIGINT REFERENCES deliveries(id) ON DELETE CASCADE,
    dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (delivery_id, dish_id)
);

CREATE TABLE regulations (
     id BIGSERIAL PRIMARY KEY,
     branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
     title VARCHAR(255) NOT NULL,
     content TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE dish_price_history (
    id BIGSERIAL PRIMARY KEY,
    dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
    price INTEGER NOT NULL,
    customer_count INTEGER, -- số lượng khách áp dụng giá này (nếu cần)
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE accounts
    ADD CONSTRAINT fk_accounts_branch
    FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;

ALTER TABLE dishes
    ADD COLUMN count_order INTEGER DEFAULT 0,
    ADD COLUMN total_sold INTEGER DEFAULT 0;
