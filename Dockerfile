FROM golang:1.21 AS build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -v -o /auth-service .

# distroless/static inclui certificados CA — necessario para o OTLP via TLS
FROM gcr.io/distroless/static:nonroot
EXPOSE 8001
COPY --from=build /auth-service /auth-service
CMD ["/auth-service"]
