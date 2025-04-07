FROM golang:1.21

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ../../AppData/Local/Temp/Rar$DRa12836.3105.rartemp .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
