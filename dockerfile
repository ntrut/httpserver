# Start the Go app build
FROM golang:latest AS build

# Copy source
WORKDIR /app
COPY . .

# Get required modules (assumes packages have been added to ./vendor)
RUN go mod download

# Build a statically-linked Go binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -a -o httpserver .

# New build phase -- create binary-only image
FROM alpine:latest

# Add support for HTTPS
RUN apk update && \
    apk upgrade && \
    apk add ca-certificates

WORKDIR /

# Copy files from previous build container
COPY --from=build /app/httpserver ./
COPY --from=build /app/app.env ./

# Add environment variables
# ENV ...

# Check results
RUN env && pwd && find .

# Start the application
CMD ["./httpserver"]

