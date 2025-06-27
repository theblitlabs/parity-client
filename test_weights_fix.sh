#!/bin/bash

# Test script to verify weights are now populated in FL results

echo "Testing federated learning weights fix..."

# Test with a simple neural network on Iris dataset
echo "Creating FL session with neural network..."
session_output=$(parity-client fl create-session-with-data ../test_data/iris.csv \
  --name "Weights Test Session" \
  --model-type neural_network \
  --total-rounds 2 \
  --min-participants 1 \
  --split-strategy random \
  --alpha 0.5 \
  --min-samples 10)

if [[ $session_output == *"Session ID:"* ]]; then
  session_id=$(echo "$session_output" | grep "Session ID:" | awk "{print \$3}")
  echo "✓ Session created successfully: $session_id"
  
  # Start the session
  echo "Starting FL session..."
  parity-client fl start-session "$session_id"
  
  # Wait for training to complete
  echo "Waiting for training to complete..."
  sleep 30
  
  # Get the trained model
  echo "Retrieving trained model..."
  model_output=$(parity-client fl get-model "$session_id")
  
  # Check if weights are present and not empty
  if [[ $model_output == *"\"weights\":"* ]] && [[ $model_output != *"\"weights\": {}"* ]]; then
    echo "✅ SUCCESS: Model weights are now properly populated!"
    echo "Weights found in model output."
  else
    echo "❌ FAILED: Model weights are still empty or missing."
    echo "Model output:"
    echo "$model_output"
  fi
else
  echo "❌ Failed to create FL session"
  echo "$session_output"
fi
