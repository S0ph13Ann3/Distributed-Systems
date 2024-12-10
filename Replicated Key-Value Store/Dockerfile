FROM golang:1.21

#Here is the working directory for the dockerfile
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and 
#only redownloading them in subsequent builds if they change
COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./...

# Explicitly expose container port 8090
EXPOSE 8090

CMD ["app"]