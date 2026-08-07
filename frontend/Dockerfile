FROM node:24-alpine AS build

ENV HUSKY=0
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .

ARG VITE_API_BASE_URL
ARG VITE_ZIG_API_BASE_URL
ARG VITE_TELEMETRY_ENABLED=false
ARG VITE_APP_ENV=production
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_ZIG_API_BASE_URL=$VITE_ZIG_API_BASE_URL
ENV VITE_TELEMETRY_ENABLED=$VITE_TELEMETRY_ENABLED
ENV VITE_APP_ENV=$VITE_APP_ENV
RUN npm run build

FROM node:24-alpine
ENV NODE_ENV=production
WORKDIR /app
COPY --from=build /app/build ./build
COPY --from=build /app/package.json ./package.json
USER node
EXPOSE 3000
CMD ["node", "build"]
