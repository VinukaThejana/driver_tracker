FROM golang:1.22.2-alpine3.19

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go mod tidy && go build -o stream

EXPOSE 8080

CMD [ "./stream" ]
