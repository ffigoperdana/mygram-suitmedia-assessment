FROM node:24-alpine AS build

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_BASE_URL
ARG VITE_USE_SAME_ORIGIN_API=false
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_USE_SAME_ORIGIN_API=${VITE_USE_SAME_ORIGIN_API}

RUN npm run build

FROM nginx:1.27-alpine

ENV API_UPSTREAM=http://api:8080
ENV NGINX_ENVSUBST_FILTER=^API_UPSTREAM$

COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -q --spider http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
