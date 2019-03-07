FROM alpine
RUN apk --update --no-cache add ca-certificates
COPY ./api /api
ENTRYPOINT ["/api"]