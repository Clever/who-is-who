FROM node:14-slim

WORKDIR /app

CMD ["npm", "start"]

COPY . /app
