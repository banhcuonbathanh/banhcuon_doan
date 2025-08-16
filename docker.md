# Remove dangling/none images

docker image prune

# To verify if the image is now local:

docker images | grep postgres

# To verify if your PostgreSQL container is running:

docker ps --filter "name=postgres"

# To get full details about this image:

docker image inspect 43eba88b28dc --format '{{.RepoTags}}'

# To get container

docker ps

# Get the actual image ID from your running container:

docker inspect --format='{{.Image}}' postgres

# Inspect using the image name (simplest method):

docker image inspect postgres:15-alpine

docker image inspect f490b1e1368a


# Pre-pull the image once to cache it locally
docker pull postgres:15-alpine

# Use --no-recreate to avoid rebuilding existing containers
docker-compose up -d --no-recreate mypostgres_ai

# Alternative: Use --pull never to never pull images
docker-compose up -d --pull never mypostgres_ai

# Check if containers are already running before starting
docker-compose ps