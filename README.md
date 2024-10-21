# URL Shortener

This project is a simple URL shortener service built with Go, demonstrating the use of the standard library's HTTP server and the Chi router. It's part of the [Rocketseat Golang course](rocketseat.com.br)

## Getting Started

To run the project:

1. Ensure you have Go installed on your system
2. Clone the repository
3. Run `go mod tidy` to install dependencies
4. Execute `go run main.go` to start the server
5. The server will be available at `http://localhost:8080`

## API Endpoints

- `POST /shorten`: Shorten a URL
- `GET /{code}`: Redirect to the original URL
