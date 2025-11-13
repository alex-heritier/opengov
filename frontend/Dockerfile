FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY frontend/package*.json /app/

# Install dependencies
RUN npm ci

# Copy source code
COPY frontend/src /app/src
COPY frontend/public /app/public
COPY frontend/index.html /app/
COPY frontend/vite.config.ts /app/
COPY frontend/tsconfig.json /app/
COPY frontend/tsconfig.app.json /app/
COPY frontend/vitest.config.ts /app/
COPY frontend/components.json /app/

# Build application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy nginx config
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built application
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
