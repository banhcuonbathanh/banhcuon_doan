1. Check if PostgreSQL is running:
   pg_isready -U restaurant -d restaurant

2. Connect to localhost (if in the same container):
   psql -h localhost -U restaurant -d restaurant

3. List all tables in the current database:
   \dt
4. View All Data
SELECT * FROM accounts;
4. 
Step 1: Run Database Migrations
First, create and run your database migrations to set up the tables:
bash# Create a new migration (if you don't have any yet)
make migrate-create

# When prompted, enter a name like: create_initial_tables
# This will create files in internal/db/migrations/

# Run the migrations to create tables
make migrate-up
brew install golang-migrate
