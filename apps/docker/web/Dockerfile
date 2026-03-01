FROM node:20-alpine AS builder

WORKDIR /build
COPY apps/node/sdk-tracker ./sdk-tracker

WORKDIR /build/web
COPY apps/node/web/package.json apps/node/web/package-lock.json* ./
RUN npm ci

COPY apps/node/web/ .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
RUN apk add --no-cache wget

COPY --from=builder /build/web/.next/standalone ./
COPY --from=builder /build/web/.next/static ./.next/static
COPY --from=builder /build/web/public ./public

ENV HOSTNAME=0.0.0.0
EXPOSE 3000
CMD ["node", "server.js"]
