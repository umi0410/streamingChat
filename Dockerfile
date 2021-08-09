FROM golang:1.16
COPY ./ /streamChat
WORKDIR /streamChat
RUN go build -o bin/streamChat
ENTRYPOINT ["bin/streamChat"]