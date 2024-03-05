FROM golang:latest


WORKDIR /scalingo

# Copy local code to the container image.
COPY . ./

RUN go mod download

RUN go install github.com/google/wire/cmd/wire@latest
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.0

# Build the binary.
RUN make

EXPOSE 5000

# Run the web service on container startup.
CMD ["/scalingo/cmd/app/app"]
