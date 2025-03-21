# Parity Protocol

Parity Protocol is a decentralized compute network where runners can execute compute tasks (e.g., running a Docker file) and earn incentives in the form of tokens. Task creators can add tasks to a pool, and the first runner to complete a task successfully receives a reward.

## Quick Start

### Prerequisites

- Go 1.22.7 or higher (using Go toolchain 1.23.4)
- Make

### Installation

1. Clone the repository:

```bash
git clone https://github.com/theblitlabs/parity-client.git
cd parity-client
```

2. Install dependencies:

```bash
make deps
```

### Setup and Running

1. Authenticate with your private key:

```bash
parity-client auth --private-key YOUR_PRIVATE_KEY
```

2. Stake tokens to participate in the network:

```bash
parity-client stake --amount 10
```

3. Run the client:

```bash
parity-client
```

### Adding Tasks

To add a new task to the network, you can use the following API endpoint:

```bash
curl -X POST http://localhost:3000/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "image": "alpine:latest",
    "command": ["echo", "Hello World"],
    "title": "Sample Task",
    "description": "This is a sample task description"
  }'
```

### Development

The project includes several helpful Makefile commands for development:

```bash
make deps          # Download dependencies
make build         # Build the application
make run           # Run the application
make clean         # Clean build files
make fmt           # Format code using gofumpt or gofmt
make imports       # Fix imports formatting
make format        # Run all formatters (gofumpt + goimports)
make lint         # Run linting
make format-lint  # Format code and run linters in one step
make watch        # Run with hot reload (requires air)
make install      # Install parity command globally
make uninstall    # Remove parity command from system
make help         # Display all available commands
```

For hot reloading during development:

```bash
# Install air (required for hot reloading)
make install-air

# Run with hot reload
make watch
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Install git hooks for automatic formatting and linting:
   ```bash
   make install-hooks
   ```
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
