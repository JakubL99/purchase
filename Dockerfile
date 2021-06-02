FROM alpine
RUN apk --no-cache add curl
ADD  purchase /purchase
ENTRYPOINT [ "/purchase" ]