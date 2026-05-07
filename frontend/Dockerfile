FROM node:20-alpine AS dev

WORKDIR /app

COPY package*.json ./

# Показываем более подробный лог, плюс явно ставим npm ci/ install
RUN npm install --loglevel=info

COPY . .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]