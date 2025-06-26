# Data Partitioning for Distributed Federated Learning

This document describes the data partitioning functionality that enables truly distributed federated learning in the PLGenesis system.

## Overview

Data partitioning is crucial for realistic federated learning scenarios where participants have different subsets of data. This implementation supports multiple partitioning strategies to simulate various real-world conditions and evaluate federated learning algorithms under different data distributions.

## Partitioning Strategies

### 1. Random (IID) Partitioning

- **Strategy**: `random` or `iid`
- **Description**: Randomly distributes data samples across participants
- **Use Case**: Baseline for federated learning with balanced data distribution
- **Parameters**:
  - `overlap_ratio`: Overlap between partitions (0.0-1.0)
  - `min_samples`: Minimum samples per participant

```bash
./parity-client fl create-session \
  --name "Random IID Test" \
  --model-type neural_network \
  --dataset-cid QmXXX... \
  --split-strategy random \
  --min-participants 3 \
  --overlap-ratio 0.0 \
  --min-samples 50
```

### 2. Stratified Partitioning

- **Strategy**: `stratified`
- **Description**: Maintains class distribution across all participants
- **Use Case**: Ensures each participant has representative samples from all classes
- **Benefits**: Balances heterogeneity while ensuring data quality

```bash
./parity-client fl create-session \
  --name "Stratified Distribution" \
  --model-type neural_network \
  --dataset-cid QmXXX... \
  --split-strategy stratified \
  --min-participants 4 \
  --min-samples 40
```

### 3. Non-IID (Dirichlet) Partitioning

- **Strategy**: `non_iid` or `dirichlet`
- **Description**: Uses Dirichlet distribution to create statistical heterogeneity
- **Use Case**: Simulates real-world federated learning with varying data distributions
- **Parameters**:
  - `alpha`: Dirichlet parameter (lower = more skewed, higher = more uniform)
  - Common values: 0.1 (highly skewed), 0.5 (moderately skewed), 1.0 (balanced)

```bash
./parity-client fl create-session \
  --name "Non-IID Heterogeneous" \
  --model-type neural_network \
  --dataset-cid QmXXX... \
  --split-strategy non_iid \
  --alpha 0.1 \
  --min-participants 5 \
  --min-samples 30
```

### 4. Label Skew Partitioning

- **Strategy**: `label_skew`
- **Description**: Each participant specializes in a subset of classes
- **Use Case**: Tests robustness when participants have completely different tasks
- **Benefits**: Simulates system heterogeneity in federated learning

```bash
./parity-client fl create-session \
  --name "Label Skew Specialization" \
  --model-type neural_network \
  --dataset-cid QmXXX... \
  --split-strategy label_skew \
  --min-participants 3 \
  --overlap-ratio 0.1 \
  --min-samples 40
```

## Configuration Parameters

### Core Parameters

- `split-strategy`: Partitioning strategy (required)
- `min-participants`: Minimum number of participants (default: 1)
- `min-samples`: Minimum samples per participant (default: 50)

### Advanced Parameters

- `alpha`: Dirichlet distribution parameter for non-IID (default: 0.5)
- `overlap-ratio`: Data overlap between participants (0.0-1.0, default: 0.0)

### Alpha Parameter Guide (for Non-IID)

- **α = 0.1**: Highly skewed, very heterogeneous data
- **α = 0.5**: Moderately skewed, realistic federated scenarios
- **α = 1.0**: Balanced distribution, approaching IID
- **α = 10.0**: Nearly uniform distribution

## Implementation Details

### Data Loader Integration

The partitioning functionality is integrated into the data loading pipeline:

```go
// Load partitioned data
partitionConfig := &training.PartitionConfig{
    Strategy:     "non_iid",
    TotalParts:   4,
    PartIndex:    1,
    Alpha:        0.5,
    MinSamples:   50,
    OverlapRatio: 0.0,
}

features, labels, err := dataLoader.LoadPartitionedData(ctx, cid, format, partitionConfig)
```

### Automatic Partition Assignment

The federated learning service automatically:

1. Determines participant list and count
2. Assigns partition indices to each participant
3. Configures partitioning parameters in training tasks
4. Distributes tasks with appropriate data subsets

### Training Task Configuration

Each training task includes partition configuration:

```json
{
  "session_id": "abc123",
  "round_id": "round1",
  "dataset_cid": "QmXXX...",
  "partition_config": {
    "strategy": "non_iid",
    "total_parts": 4,
    "part_index": 1,
    "alpha": 0.5,
    "min_samples": 50,
    "overlap_ratio": 0.0
  }
}
```

## Testing and Validation

### Test Script

Use the provided test script to evaluate different partitioning strategies:

```bash
./test_distributed_fl.sh
```

This script tests all partitioning strategies with the Iris dataset and provides comparative analysis.

### Manual Testing

Create sessions with different strategies and compare:

```bash
# Test 1: Random IID
./parity-client fl create-session --split-strategy random --alpha 1.0

# Test 2: Non-IID with high skew
./parity-client fl create-session --split-strategy non_iid --alpha 0.1

# Test 3: Label specialization
./parity-client fl create-session --split-strategy label_skew
```

## Benefits of Data Partitioning

### 1. Realistic Federated Learning

- Simulates real-world data distribution scenarios
- Tests algorithm robustness under various conditions
- Evaluates convergence with heterogeneous data

### 2. Privacy and Security

- Each participant only accesses their data subset
- No need to share or replicate full datasets
- Natural data isolation and privacy preservation

### 3. Scalability

- Supports arbitrary numbers of participants
- Configurable partition sizes and overlaps
- Efficient data distribution strategies

### 4. Research and Development

- Enables comparative studies of FL algorithms
- Benchmarking under different data conditions
- Investigation of aggregation strategies

## Performance Considerations

### Data Distribution Impact

- **IID data**: Faster convergence, similar to centralized training
- **Non-IID data**: Slower convergence, may require more rounds
- **Label skew**: Most challenging, may need specialized aggregation

### Optimization Strategies

- Increase `local_epochs` for non-IID scenarios
- Adjust `learning_rate` based on data heterogeneity
- Use `min_samples` to ensure statistical significance
- Consider `overlap_ratio` for critical applications

## Advanced Usage

### Custom Partition Strategies

Extend the `PartitionConfig` to implement custom strategies:

```go
type PartitionConfig struct {
    Strategy      string  `json:"strategy"`
    TotalParts    int     `json:"total_parts"`
    PartIndex     int     `json:"part_index"`
    Alpha         float64 `json:"alpha"`
    MinSamples    int     `json:"min_samples"`
    OverlapRatio  float64 `json:"overlap_ratio"`
    CustomParams  map[string]interface{} `json:"custom_params,omitempty"`
}
```

### Integration with Existing Datasets

The partitioning works with any dataset format supported by the system:

- CSV files with categorical or numeric labels
- JSON structured data
- IPFS/Filecoin distributed datasets

### Multi-Round Consistency

Partitions remain consistent across training rounds:

- Same participants get same data subsets
- Deterministic partitioning based on participant indices
- Maintains data locality throughout training

## Troubleshooting

### Common Issues

1. **Insufficient data**: Increase dataset size or reduce `min_samples`
2. **Unbalanced partitions**: Use `stratified` strategy for balanced classes
3. **Slow convergence**: Increase `local_epochs` for non-IID scenarios
4. **Memory issues**: Reduce `min_samples` or use sequential loading

### Validation

Monitor partition quality:

- Check sample counts per participant
- Verify class distributions
- Monitor training convergence rates
- Compare aggregation metrics

## Future Enhancements

### Planned Features

- Dynamic repartitioning during training
- Adaptive alpha parameter tuning
- Geographic-based partitioning strategies
- Temporal data partitioning support

### Community Contributions

The partitioning system is designed to be extensible. Contributions for new strategies, optimizations, and research applications are welcome.

## Conclusion

Data partitioning transforms the PLGenesis federated learning system from a simulation into a truly distributed training platform. By supporting multiple partitioning strategies, researchers and practitioners can evaluate federated learning algorithms under realistic conditions and develop robust solutions for real-world deployments.
