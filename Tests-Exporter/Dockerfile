FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum main.go ./

RUN go mod download \
    && go build -o /app/main .

FROM mcr.microsoft.com/powershell:latest

WORKDIR /app

COPY ./Lib /app/Lib
COPY ./Tests /app/Tests
COPY --from=build /app/main /app/main

EXPOSE 9099

CMD [ "pwsh", "-Command", "/app/main" ]
