FROM node:18-buster

WORKDIR /app

COPY . /app

CMD ["node", "--require", "./tracing.js", "server.js"]
