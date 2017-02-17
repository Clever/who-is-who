FROM node:6.2.2-slim

WORKDIR /app

CMD ["npm", "start"]

COPY . /app
