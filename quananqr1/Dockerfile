FROM node:21.7.1

# Build arguments with defaults
ARG NODE_ENV=development
ARG NEXT_PUBLIC_API_ENDPOINT=http://localhost:8888/
ARG NEXT_PUBLIC_URL=http://localhost:3000/

# Set environment variables
ENV NODE_ENV=$NODE_ENV
ENV NEXT_PUBLIC_API_ENDPOINT=$NEXT_PUBLIC_API_ENDPOINT
ENV NEXT_PUBLIC_URL=$NEXT_PUBLIC_URL
ENV NODE_OPTIONS='--openssl-legacy-provider'
ENV NEXT_TELEMETRY_DISABLED=1

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm install

# Copy the rest of the application
COPY . .

# Expose port
EXPOSE 3000

# Start command based on NODE_ENV
CMD if [ "$NODE_ENV" = "production" ]; then \
        npm run build && npm start; \
    else \
        npm run dev; \
    fi
# RUN npm run build

# CMD ["npm", "start"]