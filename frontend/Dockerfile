# --- build stage ---
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json ./
RUN npm install
COPY . .
# VITE_API_URL must be set at build time (Vite bakes it into the bundle)
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# --- run stage: serve with nginx ---
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# Single-page app: all routes fall back to index.html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
