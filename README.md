# Parity Client

The command-line interface for the PLGenesis decentralized AI and compute network. Parity Client provides an intuitive way to interact with the network, submit tasks, execute LLM inference, and manage your account.

## 🚀 Features

### 🤖 LLM Interaction

- **Model Discovery**: List all available LLM models across the network
- **Prompt Submission**: Submit prompts for processing with real-time status tracking
- **Async Processing**: Non-blocking prompt submission with optional wait functionality
- **Response Retrieval**: Get completed responses with comprehensive metadata

### ⚡ Task Management

- **Task Submission**: Submit compute tasks to the network
- **Status Monitoring**: Real-time tracking of task progress and completion
- **Result Retrieval**: Get task outputs and execution logs
- **Batch Operations**: Submit multiple tasks efficiently

### 🔒 Account Management

- **Authentication**: Secure authentication with private keys
- **Staking**: Stake tokens to participate as a runner
- **Balance Checking**: Monitor token balances and staking status
- **Transaction History**: View account activity and earnings

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

The client is configured using environment variables through a `.env` file. The application supports multiple config file locations for development and production use.

### Config File Locations

The client looks for configuration files in the following order:

1. **Production (after `make install`)**: `~/.parity/.env`
2. **Development/Custom path**: Specified via `--config-path` flag
3. **Local fallback**: `./.env` in the current directory

### Initial Setup

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
ETHEREUM_TOKEN_ADDRESS="0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"
ETHEREUM_STAKE_WALLET_ADDRESS="0x7465e7a637f66cb7b294b856a25bc84abff1d247"
ETHEREUM_CHAIN_ID=314159
ETHEREUM_RPC="https://api.calibration.node.glif.io/rpc/v1"
```

### Installing the Client

To install the client globally and set up the production config:

```bash
make install
```

This will:

- Install the `parity-client` binary to `/usr/local/bin`
- Create the `~/.parity` directory
- Copy your current `.env` file to `~/.parity/.env`
- If `~/.parity/.env` already exists, prompt you to confirm replacement (defaults to Yes)

After installation, you can run `parity-client` from any directory and it will automatically use the config from `~/.parity/.env`.

### Custom Config Path

You can specify a custom config path using the `--config-path` flag:

```bash
parity-client --config-path /path/to/custom.env
```

### Uninstalling

To remove the client and clean up config files:

```bash
make uninstall
```

This removes both the binary and the `~/.parity` directory.

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

### LLM Operations

#### List Available Models

See which LLM models are currently available:

```bash
parity-client llm list-models
```

#### Submit LLM Prompts

Submit a prompt for processing:

```bash
# Submit and wait for completion
parity-client llm submit --model "qwen3:latest" --prompt "Explain quantum computing" --wait

# Submit without waiting (async)
parity-client llm submit --model "llama2:7b" --prompt "Write a Python function to sort a list"

# Check status later
parity-client llm status <prompt-id>
```

#### List Recent Prompts

View your recent LLM prompts:

```bash
parity-client llm list --limit 10
```

### Adding Tasks

To add a new compute task to the network, use the REST API:

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
   - For development: Check `.env` in your project directory
   - For installed client: Check `~/.parity/.env` or run `parity-client --help` to see current config path
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
