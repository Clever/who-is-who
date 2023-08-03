FROM node:18-buster

WORKDIR /app

COPY . /app

RUN npm install --userconfig .npmrc_docker

CMD ["node", "--require", "./tracing.js", "server.js"]
