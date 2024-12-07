FROM golang:1.22 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server /server
RUN chmod +x /server
EXPOSE 8080
CMD ["/server"]