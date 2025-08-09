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
