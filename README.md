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

### Development

The project includes several helpful Makefile commands for development:

```bash
make build          # Build the application
make run           # Run the application
make clean         # Clean build files
make deps          # Download dependencies
make fmt           # Format code
make watch         # Run with hot reload (requires air)
make install       # Install parity command globally
make uninstall     # Remove parity command from system
make help          # Display all available commands
```

For hot reloading during development:

```bash
# Install air (required for hot reloading)
make install-air

# Run with hot reload
make watch
```

### CLI Commands

The CLI provides a unified interface through the `parity` command:

```bash
# Show available commands and help
parity help

# Authenticate with your private key
parity auth --private-key <private-key>

# Check balance
parity balance

# Stake tokens
parity stake --amount <amount>

# Start chain proxy
parity chain
```

Each command supports the `--help` flag for detailed usage information:

```bash
parity <command> --help
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
