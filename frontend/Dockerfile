# Build stage
FROM node:24-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm install

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Final stage - just copy built files to volume
FROM alpine:latest

WORKDIR /app

# Copy built files
COPY --from=builder /app/dist ./dist

# Copy built files to volume on startup, then keep container alive
CMD ["sh", "-c", "cp -r /app/dist/. /app/volume/ && echo 'Frontend build completed' && sleep infinity"]
