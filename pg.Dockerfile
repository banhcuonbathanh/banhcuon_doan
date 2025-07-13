FROM postgres:15-alpine

# Set environment variables
ENV POSTGRES_DB=restaurant \
    POSTGRES_USER=restaurant \
    POSTGRES_PASSWORD=restaurant

# The official postgres image automatically handles directory creation and permissions
# No need to manually create directories

EXPOSE 5432