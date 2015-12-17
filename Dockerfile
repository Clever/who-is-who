FROM debian:jessie
RUN apt-get update && apt-get install -y ca-certificates
COPY bin/who-is-who  /usr/bin/who-is-who
CMD ["who-is-who"]
