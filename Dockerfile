FROM golang:1.21.3-bullseye AS builder

# Env setup
ENV CGO_ENABLED="0"
# Setup workdir
WORKDIR /build
# Copy source code
COPY . .

# Fetch dependencies
RUN go mod download

RUN GOOS=linux go build -ldflags="-s -w" -o app .

# Runner stage
FROM debian:bullseye-slim AS runner

# Build Args
ARG BINARY_NAME="app"
ARG START_COMMAND="./app"

# Setup workdir
WORKDIR /user

# Copy binary
COPY --from=builder /build/app .

# Create entrypoint
RUN echo "./app" > /user/entrypoint.sh
RUN chmod +x /user/entrypoint.sh

# Setup Entrypoint
ENTRYPOINT ["sh", "-c", "/user/entrypoint.sh"]