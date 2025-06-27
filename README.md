# Parity Client

The command-line interface for the PLGenesis decentralized AI and compute network. Parity Client provides an intuitive way to interact with the network, submit tasks, execute LLM inference, and manage federated learning sessions.

## 🚀 Features

### 🤖 LLM Interaction

- **Model Discovery**: List all available LLM models across the network
- **Prompt Submission**: Submit prompts for processing with real-time status tracking
- **Async Processing**: Non-blocking prompt submission with optional wait functionality
- **Response Retrieval**: Get completed responses with comprehensive metadata

### 🧠 Federated Learning

- **Session Management**: Create and manage distributed federated learning sessions
- **Data Partitioning**: Support for 5 partitioning strategies (random, stratified, sequential, non-IID, label skew)
- **Model Training**: Coordinate distributed training across multiple participants
- **Model Aggregation**: Retrieve trained models from completed sessions
- **Real-time Monitoring**: Track training progress and participant status

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
  - [Federated Learning](#federated-learning)
  - [LLM Operations](#llm-operations)
  - [Task Management](#task-management)
- [Configuration Files](#configuration-files)
- [API Reference](#api-reference)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
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
# Server Configuration
SERVER_HOST="0.0.0.0"
SERVER_PORT=3000
SERVER_ENDPOINT="/api"

# Filecoin Network Configuration (Filecoin Calibration Testnet)
FILECOIN_RPC=https://calibration.filfox.info/rpc/v1
FILECOIN_CHAIN_ID=314159
FILECOIN_TOKEN_ADDRESS=0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0
FILECOIN_STAKE_WALLET_ADDRESS=0x1fd2fA91C1eE96d128A35955EC3B496D7184c80A

# IPFS/Filecoin Storage Configuration
FILECOIN_IPFS_ENDPOINT=http://localhost:5001
FILECOIN_GATEWAY_URL=https://gateway.pinata.cloud
FILECOIN_CREATE_STORAGE_DEALS=false

# Runner Configuration
RUNNER_SERVER_URL="http://localhost:8080"
RUNNER_WEBHOOK_PORT=8082
RUNNER_API_PREFIX="/api"

# Federated Learning Configuration
FL_SERVER_URL="http://localhost:8080"
FL_DEFAULT_TIMEOUT=30s
FL_RETRY_ATTEMPTS=3
FL_LOG_LEVEL=info
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

### Federated Learning

The federated learning system requires explicit configuration for all parameters. No default values are used to ensure complete transparency and control.

#### Creating Model Configuration Files

Before creating FL sessions, you need to prepare model configuration files:

**neural_network_config.json**:

```json
{
  "input_size": 784,
  "hidden_size": 128,
  "output_size": 10
}
```

**linear_regression_config.json**:

```json
{
  "input_size": 13,
  "output_size": 1
}
```

#### 1. Create Federated Learning Session

```bash
# Basic neural network session
parity-client fl create-session \
  --name "MNIST Classification" \
  --description "Distributed MNIST digit classification" \
  --model-type neural_network \
  --total-rounds 10 \
  --min-participants 3 \
  --dataset-cid QmYourDatasetCID \
  --config-file neural_network_config.json \
  --aggregation-method federated_averaging \
  --learning-rate 0.001 \
  --batch-size 32 \
  --local-epochs 5 \
  --split-strategy random \
  --min-samples 100 \
  --alpha 0.5 \
  --overlap-ratio 0.0

# Non-IID partitioning session
parity-client fl create-session \
  --name "Non-IID Training" \
  --model-type neural_network \
  --total-rounds 5 \
  --dataset-cid QmYourDatasetCID \
  --config-file neural_network_config.json \
  --aggregation-method federated_averaging \
  --learning-rate 0.005 \
  --batch-size 64 \
  --local-epochs 3 \
  --split-strategy non_iid \
  --alpha 0.1 \
  --min-samples 50 \
  --overlap-ratio 0.0
```

#### 2. Upload Data and Create Session in One Step

```bash
parity-client fl create-session-with-data ./mnist_dataset.csv \
  --name "MNIST Training" \
  --model-type neural_network \
  --total-rounds 10 \
  --min-participants 2 \
  --config-file neural_network_config.json \
  --aggregation-method federated_averaging \
  --learning-rate 0.001 \
  --batch-size 32 \
  --local-epochs 5 \
  --split-strategy stratified \
  --min-samples 100 \
  --alpha 0.5 \
  --overlap-ratio 0.0
```

#### 3. Session Management

```bash
# List all sessions
parity-client fl list-sessions

# List sessions by creator
parity-client fl list-sessions --creator-address 0x123...

# Get detailed session info
parity-client fl get-session SESSION_ID

# Start a session
parity-client fl start-session SESSION_ID
```

#### 4. Model Retrieval

```bash
# Display trained model
parity-client fl get-model SESSION_ID

# Save model to file
parity-client fl get-model SESSION_ID --output trained_model.json
```

#### 5. Manual Model Updates (Advanced)

```bash
parity-client fl submit-update \
  --session-id SESSION_ID \
  --round-id ROUND_ID \
  --runner-id RUNNER_ID \
  --gradients-file gradients.json \
  --data-size 1000 \
  --loss 0.25 \
  --accuracy 0.85
```

#### Data Partitioning Strategies

The system supports five partitioning strategies:

1. **Random (IID)**: `--split-strategy random`

   - Uniform random distribution
   - Requires: `--min-samples`

2. **Stratified**: `--split-strategy stratified`

   - Maintains class distribution
   - Requires: `--min-samples`

3. **Sequential**: `--split-strategy sequential`

   - Consecutive data splits
   - Requires: `--min-samples`

4. **Non-IID**: `--split-strategy non_iid`

   - Dirichlet distribution for class imbalance
   - Requires: `--alpha`, `--min-samples`
   - Lower alpha = more skewed distribution

5. **Label Skew**: `--split-strategy label_skew`
   - Each participant gets subset of classes
   - Requires: `--min-samples`
   - Optional: `--overlap-ratio` for class overlap

### LLM Operations

#### List Available Models

```bash
parity-client llm list-models
```

#### Submit LLM Prompts

```bash
# Submit and wait for completion
parity-client llm submit --model "qwen3:latest" --prompt "Explain quantum computing" --wait

# Submit without waiting (async)
parity-client llm submit --model "llama2:7b" --prompt "Write a Python function to sort a list"

# Check status later
parity-client llm status <prompt-id>
```

#### List Recent Prompts

```bash
parity-client llm list --limit 10
```

### Task Management

Submit compute tasks to the network:

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

## Configuration Files

### Model Configuration Examples

#### Neural Network for MNIST

```json
{
  "input_size": 784,
  "hidden_size": 128,
  "output_size": 10
}
```

#### Neural Network for CIFAR-10

```json
{
  "input_size": 3072,
  "hidden_size": 256,
  "output_size": 10
}
```

#### Linear Regression for Boston Housing

```json
{
  "input_size": 13,
  "output_size": 1
}
```

#### Large Neural Network

```json
{
  "input_size": 2048,
  "hidden_size": 512,
  "output_size": 100
}
```

### Differential Privacy Configuration

Enable differential privacy in your sessions:

```bash
parity-client fl create-session \
  --name "Private Training" \
  --model-type neural_network \
  --config-file neural_network_config.json \
  --enable-differential-privacy \
  --noise-multiplier 0.1 \
  --l2-norm-clip 1.0 \
  # ... other required flags
```

## API Reference

### Federated Learning Endpoints

| Method | Endpoint                                       | Description          |
| ------ | ---------------------------------------------- | -------------------- |
| POST   | /api/v1/federated-learning/sessions            | Create FL session    |
| GET    | /api/v1/federated-learning/sessions            | List FL sessions     |
| GET    | /api/v1/federated-learning/sessions/{id}       | Get session details  |
| POST   | /api/v1/federated-learning/sessions/{id}/start | Start FL session     |
| GET    | /api/v1/federated-learning/sessions/{id}/model | Get trained model    |
| POST   | /api/v1/federated-learning/model-updates       | Submit model updates |

### LLM Endpoints

| Method | Endpoint                | Description                        |
| ------ | ----------------------- | ---------------------------------- |
| GET    | `/api/llm/models`       | List all available LLM models      |
| POST   | `/api/llm/prompts`      | Submit a prompt for LLM processing |
| GET    | `/api/llm/prompts/{id}` | Get prompt status and response     |
| GET    | `/api/llm/prompts`      | List recent prompts                |

### Task Endpoints

| Method | Endpoint               | Description      |
| ------ | ---------------------- | ---------------- |
| POST   | /api/tasks             | Create task      |
| GET    | /api/tasks             | List all tasks   |
| GET    | /api/tasks/{id}        | Get task details |
| GET    | /api/tasks/{id}/status | Get task status  |
| GET    | /api/tasks/{id}/logs   | Get task logs    |

### Storage Endpoints

| Method | Endpoint                    | Description                  |
| ------ | --------------------------- | ---------------------------- |
| POST   | /api/storage/upload         | Upload file to IPFS/Filecoin |
| GET    | /api/storage/download/{cid} | Download file by CID         |
| GET    | /api/storage/info/{cid}     | Get file information         |
| POST   | /api/storage/pin/{cid}      | Pin file to IPFS             |

### Runner Endpoints

| Method | Endpoint               | Description     |
| ------ | ---------------------- | --------------- |
| POST   | /api/runners           | Register runner |
| POST   | /api/runners/heartbeat | Send heartbeat  |

### Health & Status Endpoints

| Method | Endpoint    | Description   |
| ------ | ----------- | ------------- |
| GET    | /api/health | Health check  |
| GET    | /api/status | System status |

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

### Common Issues

1. **Configuration Issues**

   - Ensure your `.env` file exists and is properly configured
   - For development: Check `.env` in your project directory
   - For installed client: Check `~/.parity/.env`
   - Verify all required environment variables are set

2. **Federated Learning Issues**

   - **Missing required flags**: All FL parameters must be explicitly provided
   - **Invalid model config**: Ensure your model configuration JSON is valid
   - **Partition validation errors**: Check strategy-specific requirements:
     - `non_iid` requires positive `--alpha` value
     - All strategies require positive `--min-samples`
   - **Learning rate errors**: Must be between 0 and 1.0

3. **Connection Issues**

   - Ensure your Filecoin RPC URL is correct and accessible
   - Check your internet connection and firewall settings
   - Verify FL server URL in configuration

4. **Authentication Errors**
   - Verify your private key is correct
   - Ensure you have sufficient tokens for staking

### Error Examples and Solutions

**Error**: `model configuration is required - please provide via --config-file flag`
**Solution**: Create a model configuration JSON file and use `--config-file`

**Error**: `alpha parameter is required for non_iid partitioning strategy`
**Solution**: Provide `--alpha` parameter when using `--split-strategy non_iid`

**Error**: `learning rate must be positive, got 0.000000`
**Solution**: Provide a positive learning rate: `--learning-rate 0.001`

**Error**: `training configuration is incomplete`
**Solution**: Ensure all required training parameters are provided:

- `--learning-rate`
- `--batch-size`
- `--local-epochs`

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
