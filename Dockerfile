FROM node:18-buster

WORKDIR /app

CMD ["npm", "start"]

COPY . /app
