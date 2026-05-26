FROM golang:1.21 AS build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -v -o /auth-service .

FROM scratch
EXPOSE 8001
COPY --from=build /auth-service /auth-service
CMD ["/auth-service"]
