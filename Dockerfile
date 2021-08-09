FROM golang:1.16
WORKDIR /streamingChat
COPY go.mod go.mod
COPY go.sum go.sum
RUN ls -al && go mod download
COPY ./ /streamingChat
RUN go build -o bin/streamingChat
ENTRYPOINT ["bin/streamingChat"]