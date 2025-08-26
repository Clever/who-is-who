FROM node:24-bullseye

WORKDIR /app

COPY . /app

CMD ["node", "--require", "./tracing.js", "server.js"]
