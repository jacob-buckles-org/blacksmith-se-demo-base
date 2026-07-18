# Anvil Analytics dashboard
FROM node:22

WORKDIR /app

COPY . .

RUN npm ci
RUN npm run build

EXPOSE 4173

CMD ["npx", "vite", "preview", "--host", "0.0.0.0", "--port", "4173"]
