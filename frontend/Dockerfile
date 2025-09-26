# Build stage
FROM node:18-alpine AS builder

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

# This container just provides the built files via volume
CMD ["sh", "-c", "echo 'Frontend build completed' && sleep infinity"]