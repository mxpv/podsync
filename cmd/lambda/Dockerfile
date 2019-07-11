FROM alpine
RUN apk --update --no-cache add ca-certificates
COPY ./lambda /app/lambda
ENTRYPOINT ["/app/lambda"]
