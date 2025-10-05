1. Check if PostgreSQL is running:
   pg_isready -U restaurant -d restaurant

2. Connect to localhost (if in the same container):
   psql -h localhost -U restaurant -d restaurant

3. List all tables in the current database:
   \dt
4. View All Data
   SELECT _ FROM accounts;
SELECT * FROM refresh_tokens;


5. Step 1: Run Database Migrations
   First, create and run your database migrations to set up the tables:
   bash# Create a new migration (if you don't have any yet)
   make migrate-create

# When prompted, enter a name like: create_initial_tables

# This will create files in internal/db/migrations/

# Run the migrations to create tables

make migrate-up
brew install golang-migrate

5. delete all data

-- First, drop all tables in the correct order (respecting foreign key dependencies)
DROP TABLE IF EXISTS delivery_dishes CASCADE;
DROP TABLE IF EXISTS deliveries CASCADE;
DROP TABLE IF EXISTS order_sets CASCADE;
DROP TABLE IF EXISTS order_dishes CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS guests CASCADE;
DROP TABLE IF EXISTS tables CASCADE;
DROP TABLE IF EXISTS set_dishes CASCADE;
DROP TABLE IF EXISTS sets CASCADE;
DROP TABLE IF EXISTS dishes CASCADE;
DROP TABLE IF EXISTS dish_price_history CASCADE;
DROP TABLE IF EXISTS regulations CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS branches CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;

-- Drop the enum type as well
DROP TYPE IF EXISTS table_status CASCADE;

6. create data base
   =-=-=-=-=-=-===-=-=-=--=-=-=-==-=-=-=-=-=-=-------=-==-=-=- start

-- Create all tables in the correct order (respecting foreign key dependencies)

-- 1. Create branches table first (no dependencies)
CREATE TABLE branches (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(255) NOT NULL,
address VARCHAR(255) NOT NULL,
phone VARCHAR(50),
manager_id BIGINT,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 2. Create accounts table (references branches but we'll add constraint later)
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
status VARCHAR(50) DEFAULT 'active',
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 3. Create dishes table
CREATE TABLE dishes (
id BIGSERIAL PRIMARY KEY,
branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
name VARCHAR(255) NOT NULL,
price INTEGER NOT NULL,
description TEXT,
image VARCHAR(255),
status VARCHAR(50) DEFAULT 'Available',
count_order INTEGER DEFAULT 0,
total_sold INTEGER DEFAULT 0,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 4. Create sets table
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
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 5. Create set_dishes junction table
CREATE TABLE set_dishes (
set_id BIGINT REFERENCES sets(id) ON DELETE CASCADE,
dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
quantity INTEGER NOT NULL,
PRIMARY KEY (set_id, dish_id)
);

-- 6. Create enum type for table status
CREATE TYPE table_status AS ENUM ('AVAILABLE', 'OCCUPIED', 'RESERVED', 'OUT_OF_SERVICE', 'TAKE_AWAY');

-- 7. Create tables table
CREATE TABLE tables (
number INTEGER PRIMARY KEY,
branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
capacity INTEGER NOT NULL,
status table_status DEFAULT 'AVAILABLE',
token VARCHAR(255) NOT NULL,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 8. Create guests table
CREATE TABLE guests (
id BIGSERIAL PRIMARY KEY,
branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
name VARCHAR(255) NOT NULL,
table_number INTEGER REFERENCES tables(number) ON DELETE SET NULL,
refresh_token VARCHAR(255),
refresh_token_expires_at TIMESTAMP WITH TIME ZONE,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 9. Create orders table
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
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 10. Create order_dishes junction table
CREATE TABLE order_dishes (
order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
quantity INTEGER NOT NULL,
PRIMARY KEY (order_id, dish_id)
);

-- 11. Create order_sets junction table
CREATE TABLE order_sets (
order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
set_id BIGINT REFERENCES sets(id) ON DELETE CASCADE,
quantity INTEGER NOT NULL,
PRIMARY KEY (order_id, set_id)
);

-- 12. Create deliveries table
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
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 13. Create delivery_dishes junction table
CREATE TABLE delivery_dishes (
delivery_id BIGINT REFERENCES deliveries(id) ON DELETE CASCADE,
dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
quantity INTEGER NOT NULL,
PRIMARY KEY (delivery_id, dish_id)
);

-- 14. Create regulations table
CREATE TABLE regulations (
id BIGSERIAL PRIMARY KEY,
branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
title VARCHAR(255) NOT NULL,
content TEXT NOT NULL,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- 15. Create dish_price_history table
CREATE TABLE dish_price_history (
id BIGSERIAL PRIMARY KEY,
dish_id BIGINT REFERENCES dishes(id) ON DELETE CASCADE,
price INTEGER NOT NULL,
customer_count INTEGER,
start_time TIMESTAMP WITH TIME ZONE NOT NULL,
end_time TIMESTAMP WITH TIME ZONE,
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Add foreign key constraints that couldn't be added during table creation
ALTER TABLE accounts
ADD CONSTRAINT fk_accounts_branch
FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;

ALTER TABLE branches
ADD CONSTRAINT fk_branches_manager
FOREIGN KEY (manager_id) REFERENCES accounts(id) ON DELETE SET NULL;

-- Create indexes for deleted_at columns (for soft delete performance)
CREATE INDEX idx_accounts_deleted_at ON accounts(deleted_at);
CREATE INDEX idx_branches_deleted_at ON branches(deleted_at);
CREATE INDEX idx_dishes_deleted_at ON dishes(deleted_at);
CREATE INDEX idx_sets_deleted_at ON sets(deleted_at);
CREATE INDEX idx_tables_deleted_at ON tables(deleted_at);
CREATE INDEX idx_guests_deleted_at ON guests(deleted_at);
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at);
CREATE INDEX idx_deliveries_deleted_at ON deliveries(deleted_at);
CREATE INDEX idx_regulations_deleted_at ON regulations(deleted_at);
CREATE INDEX idx_dish_price_history_deleted_at ON dish_price_history(deleted_at);

    CREATE TABLE IF NOT EXISTS refresh_tokens (
        id BIGSERIAL PRIMARY KEY,
        account_id BIGINT NOT NULL,
        token TEXT NOT NULL UNIQUE,
        expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        revoked_at TIMESTAMP WITH TIME ZONE,
        is_revoked BOOLEAN DEFAULT FALSE,
        CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_refresh_tokens_account_id ON refresh_tokens(account_id);
    CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
    CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
    CREATE INDEX IF NOT EXISTS idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked) WHERE is_revoked = FALSE;

-==-=-=-=--=-=-=-==-=-=-=-=-=-==-=-=-=-=-=-=-=-=-==-=-=-=-=-=-==- end

Perfect! You have successfully created all 14 tables. Now let's verify the relationships and table structures:

## Verify Table Relationships and Structure

### 1. Check the accounts table structure (most important for your Go struct):

```sql
\d accounts
```

### 2. Check all foreign key relationships:

```sql
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    tc.constraint_name
FROM
    information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
ORDER BY tc.table_name, kcu.column_name;
```

### 3. Check specific table relationships:

```sql
-- Check accounts table relationships
\d accounts

-- Check branches table relationships
\d branches

-- Check orders table relationships
\d orders

-- Check deliveries table relationships
\d deliveries
```

### 4. Verify indexes were created:

```sql
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
AND indexname LIKE '%deleted_at%'
ORDER BY tablename;
```

### 5. Verify the custom enum type:

```sql
\dT table_status
```

### 6. Test that the accounts table matches your Go struct:

```sql
SELECT column_name, data_type, column_default, is_nullable
FROM information_schema.columns
WHERE table_name = 'accounts'
ORDER BY ordinal_position;
```

**Copy and paste these verification commands one by one into your terminal to check everything is properly set up!**

The key things to verify:

- ✅ All 14 tables created
- ✅ `accounts` table has `status` column with default 'active'
- ✅ All tables have `deleted_at` column for soft deletes
- ✅ Foreign key relationships are properly established
- ✅ Indexes on `deleted_at` columns exist for performance

Let me know what each verification command shows!

### Method 1: Check All Tables in public Schema (Quick & Common)

SELECT
tablename,
(xpath('/row/cnt/text()',
query_to_xml(format('SELECT COUNT(\*) AS cnt FROM %I', tablename), false, true, '')))[1]::text::int AS row_count
FROM
pg_tables
WHERE
schemaname = 'public'
ORDER BY
tablename;

# DELETE — Slower but safer in complex cases

DELETE FROM accounts;

# View Foreign Keys for All Tables

SELECT
tc.table_name,
kcu.column_name,
ccu.table_name AS foreign_table_name,
ccu.column_name AS foreign_column_name
FROM
information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
ON tc.constraint_name = kcu.constraint_name
AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
ON ccu.constraint_name = tc.constraint_name
AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
ORDER BY tc.table_name;

# INSERT INTO branches

INSERT INTO branches (id, name, status, created_at, updated_at)
VALUES (1, 'Main Branch', 'active', NOW(), NOW());
