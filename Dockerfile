FROM node:14-buster

WORKDIR /app

CMD ["npm", "start"]

COPY . /app
