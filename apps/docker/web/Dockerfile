FROM node:20-alpine AS builder

WORKDIR /app
COPY apps/node/web/package.json apps/node/web/package-lock.json* ./
RUN npm ci

COPY apps/node/web/ .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
RUN apk add --no-cache wget

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

ENV HOSTNAME=0.0.0.0
EXPOSE 3000
CMD ["node", "server.js"]
