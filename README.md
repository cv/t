# t

A quick and simple world clock for the command-line using IATA airport codes.

## Installation

```bash
go install github.com/cv/t/cmd/t@latest
```

## Usage

```bash
$ t sfo jfk
SFO: 🕓  16:06:21 (America/Los_Angeles)
JFK: 🕖  19:06:21 (America/New_York)
```

Any IATA airport code can be used, and will pick the timezone of that airport.

### Shell Prompt Mode

If `PS1_FORMAT` is set, the output will be compact with no decorations or newline, suitable for shell prompts:

```bash
$ echo $(PS1_FORMAT=1 t sfo lon)
SFO 17:47 LON 01:47
```

## Development

### Prerequisites

- Go 1.21 or later

### Building

```bash
go build -o t ./cmd/t
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
golangci-lint run
```

## Project Structure

```
t/
├── cmd/t/              # Main application entry point
│   └── main.go
├── codes/              # IATA airport code to timezone mapping
│   └── iata.go
├── internal/clock/     # Core clock display logic
│   ├── clock.go
│   └── clock_test.go
├── go.mod
├── go.sum
└── README.md
```

## License

MIT License

Copyright 2017 Carlos Villela <cv@lixo.org>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
