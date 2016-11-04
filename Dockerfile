FROM gliderlabs/alpine:3.3

RUN apk-install ca-certificates
COPY build/who-is-who /bin/who-is-who

CMD ["/bin/who-is-who", "--addr=0.0.0.0:80"]

