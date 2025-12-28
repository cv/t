# t

[![CI](https://github.com/cv/t/actions/workflows/ci.yml/badge.svg)](https://github.com/cv/t/actions/workflows/ci.yml)

A quick and simple world clock for the command-line using IATA airport codes.

## Installation

### Homebrew

```bash
brew install cv/tap/t
```

### From source

```bash
go install github.com/cv/t/cmd/t@latest
```

### From releases

Download the latest binary from the [releases page](https://github.com/cv/t/releases).

## Usage

```bash
$ t sfo jfk
SFO: 🕓  16:06:21 (America/Los_Angeles)
JFK: 🕖  19:06:21 (America/New_York)
```

Any IATA airport code can be used, and will pick the timezone of that airport.

### Version

```bash
$ t --version
t v1.0.0 (commit: abc1234, built: 2024-01-01T00:00:00Z)
```

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
make test

# Run with coverage
make test-cover

# Run with race detector
make test-race
```

### Linting

```bash
make lint
```

### Releasing

Releases are automated via GitHub Actions. To create a new release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Project Structure

```
t/
├── .github/workflows/  # CI and release automation
│   ├── ci.yml
│   └── release.yml
├── cmd/t/              # Main application entry point
│   └── main.go
├── codes/              # IATA airport code to timezone mapping
│   └── iata.go
├── internal/clock/     # Core clock display logic
│   ├── clock.go
│   └── clock_test.go
├── .goreleaser.yml     # Release configuration
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
