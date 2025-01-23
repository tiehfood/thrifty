FROM node:alpine AS builder

COPY . /app
WORKDIR /app

RUN wget -qO- https://get.pnpm.io/install.sh | ENV="$HOME/.shrc" SHELL="$(which sh)" sh -; \
    source /root/.shrc; \
    pnpm install; \
    pnpm build;

FROM nginxinc/nginx-unprivileged:alpine3.20

COPY --from=builder /app/build /usr/share/nginx/html
USER nginx
