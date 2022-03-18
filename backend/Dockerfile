FROM golang:alpine

RUN mkdir /app

COPY go.mod go.sum /app/

WORKDIR /app

RUN go mod download

COPY .* *.go /app/

RUN go build -o backend *.go

# RUn go run .

CMD [ "/app/backend" ]
