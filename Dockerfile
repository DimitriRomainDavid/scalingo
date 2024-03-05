FROM golang:latest

WORKDIR /scalingo

# Copy local code to the container image.
COPY . ./
RUN go mod download

# Build the binary.
RUN go build -C cmd/app/ -v -o app && mv cmd/app/app .

EXPOSE 5000

# Run the web service on container startup.
CMD ["/scalingo/app"]
