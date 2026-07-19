FROM node:20-alpine AS build

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_BASE_URL
ARG VITE_USE_SAME_ORIGIN_API=false
ARG VITE_CAP_ENABLED=false
ARG VITE_CAP_BASE_URL=
ARG VITE_CAP_SITE_KEY=
ARG VITE_CAP_REQUIRED_ON_LOGIN=true
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_USE_SAME_ORIGIN_API=${VITE_USE_SAME_ORIGIN_API}
ENV VITE_CAP_ENABLED=${VITE_CAP_ENABLED}
ENV VITE_CAP_BASE_URL=${VITE_CAP_BASE_URL}
ENV VITE_CAP_SITE_KEY=${VITE_CAP_SITE_KEY}
ENV VITE_CAP_REQUIRED_ON_LOGIN=${VITE_CAP_REQUIRED_ON_LOGIN}

RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -q --spider http://127.0.0.1/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
