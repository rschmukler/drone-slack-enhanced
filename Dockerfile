FROM gliderlabs/alpine:3.1
RUN apk-install ca-certificates
ADD drone-slack /bin/
ENTRYPOINT ["/bin/drone-slack"]
