FROM node:20.19.0-alpine AS base
WORKDIR /app

# Копируем package.json и package-lock.json для кэширования зависимостей
COPY package.json package-lock.json ./

# Устанавливаем зависимости
RUN npm ci

# Продуктивная сборка
FROM base AS prod
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0"]

# Для разработки — копируем код, но не перезатираем node_modules
FROM base AS dev
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]