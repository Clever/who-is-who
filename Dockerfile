FROM debian:jessie
COPY bin/who-is-who  /usr/bin/who-is-who
CMD ["who-is-who"]
