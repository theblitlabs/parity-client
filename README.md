# Parity Protocol

Parity Protocol is a decentralized compute network that enables distributed task execution with blockchain-based incentives. The platform allows:

- **Task Creators** to submit compute tasks to the network
- **Runners** to execute tasks and earn token rewards
- **Secure Execution** of Docker containers and compute workloads
- **Decentralized Verification** of task completion
- **Token-based Incentives** for participating in the network

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Setup and Running](#setup-and-running)
  - [Adding Tasks](#adding-tasks)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)

## Quick Start

### Prerequisites

#### Software Requirements

- Go 1.22.7 or higher (using Go toolchain 1.23.4)
- Make
- Docker (latest version recommended)

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

3. Build the client:

```bash
make build
```

## Configuration

The client is configured using environment variables through a `.env` file:

1. Copy the example environment file:

```bash
cp .env.sample .env
```

2. Edit `.env` with your settings:

```bash
# Runner config
RUNNER_WEBHOOK_PORT=8090
RUNNER_API_PREFIX="/api"
RUNNER_SERVER_URL="http://localhost:8080/api"

# Server config
SERVER_HOST="0.0.0.0"
SERVER_PORT=3000
SERVER_ENDPOINT="/api"

# Ethereum config
ETHEREUM_TOKEN_ADDRESS="0x..."
ETHEREUM_STAKE_WALLET_ADDRESS="0x..."
ETHEREUM_CHAIN_ID=11155111
ETHEREUM_RPC="https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY"
```

You can also specify a custom config path using the `--config-path` flag:

```bash
parity-client --config-path /path/to/custom.env
```

## Usage

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

To add a new task to the network, use the REST API:

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

## Development

The project includes several helpful Makefile commands for development:

```bash
make deps          # Download dependencies
make build         # Build the application
make run           # Run the application
make clean         # Clean build files
make fmt          # Format code using gofumpt or gofmt
make imports       # Fix imports formatting
make format       # Run all formatters (gofumpt + goimports)
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

## Troubleshooting

Common issues and solutions:

1. **Configuration Issues**

   - Ensure your `.env` file exists and is properly configured
   - Check that all required environment variables are set
   - Verify the config path if using `--config-path`

2. **Connection Issues**

   - Ensure your Ethereum RPC URL is correct and accessible
   - Check your internet connection
   - Verify your firewall settings

3. **Authentication Errors**

   - Verify your private key is correct
   - Ensure you have sufficient tokens for staking

4. **Task Execution Failures**
   - Check Docker is running and accessible
   - Verify you have sufficient disk space
   - Ensure required ports are not blocked

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Install git hooks for automatic formatting and linting:
   ```bash
   make install-hooks
   ```
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

Please ensure your PR:

- Follows the existing code style
- Includes appropriate tests
- Updates documentation as needed
- Describes the changes in detail

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
