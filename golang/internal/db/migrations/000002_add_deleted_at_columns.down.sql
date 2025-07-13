-- Drop indexes for deleted_at columns
DROP INDEX IF EXISTS idx_accounts_deleted_at;
DROP INDEX IF EXISTS idx_branches_deleted_at;
DROP INDEX IF EXISTS idx_dishes_deleted_at;
DROP INDEX IF EXISTS idx_sets_deleted_at;
DROP INDEX IF EXISTS idx_tables_deleted_at;
DROP INDEX IF EXISTS idx_guests_deleted_at;
DROP INDEX IF EXISTS idx_orders_deleted_at;
DROP INDEX IF EXISTS idx_deliveries_deleted_at;
DROP INDEX IF EXISTS idx_regulations_deleted_at;
DROP INDEX IF EXISTS idx_dish_price_history_deleted_at;

-- Remove deleted_at column from accounts table
ALTER TABLE accounts 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from branches table
ALTER TABLE branches 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from dishes table
ALTER TABLE dishes 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from sets table
ALTER TABLE sets 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from tables table
ALTER TABLE tables 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from guests table
ALTER TABLE guests 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from orders table
ALTER TABLE orders 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from deliveries table
ALTER TABLE deliveries 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from regulations table
ALTER TABLE regulations 
DROP COLUMN IF EXISTS deleted_at;

-- Remove deleted_at column from dish_price_history table
ALTER TABLE dish_price_history 
DROP COLUMN IF EXISTS deleted_at; 