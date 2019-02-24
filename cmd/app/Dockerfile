FROM alpine
RUN apk --update --no-cache add ca-certificates
COPY ./server /server
ENTRYPOINT ["/server"]