#!/bin/bash

set -e

# Define and setup config and data files if they don't exist.
config_file="$HOME/.config/triggerpubsub/last_command.json"
data_file="$HOME/.config/triggerpubsub/last_data.txt"

if [[ ! -f "$config_file" ]]; then
  mkdir -p "$(dirname "$config_file")"
  echo "{}" > "$config_file"
fi
if [[ ! -f "$data_file" ]]; then
  mkdir -p "$(dirname "$data_file")"
  touch "$data_file"
fi

# Get previous GCP project ID and PubSub topic to publish to from config file.
PROJECT_ID=""
PROJECT_ID=$(jq -r .project "$config_file")
if [[ $PROJECT_ID == "null" ]]; then
  PROJECT_ID=""
fi

TOPIC=$(jq -r .topic "$config_file")
if [[ $TOPIC == "null" ]]; then
  TOPIC=""
fi

# Get previous data published.
DATA=$(cat "$data_file")

# Enter project ID to publish a topic from.
echo "Enter a project ($PROJECT_ID):"

read -r input

if [[ ! $input == "" ]]; then
   PROJECT_ID="$input"
fi

# Enter topic to publish a message to.
echo "Enter a topic to publish to ($TOPIC):"

read -r input

if [[ ! $input == "" ]]; then
   TOPIC="$input"
fi

# Enter data to publish. Accepts multiline input, and needs to be json format.
echo "Enter data to publish:"
echo "Default"
echo "$DATA" | jq .

# shellcheck disable=2162
read input

if [[ ! $input == "" ]]; then
   DATA="$input"
fi

# Verify the message to publish.
echo "Publishing the following message"
echo "Project: $PROJECT_ID"
echo "Topic: $TOPIC"
echo "Data:"
echo "$DATA" | jq .

echo "Is this ok? (y/n)"

read -r input

# Update the latest message sent into the config and data file.
update_history() {
  result=$(jq ".topic = \"$TOPIC\" | .project = \"$PROJECT_ID\"" "$config_file")
  echo "$result" > "$config_file"

  echo "$DATA" > "$data_file"
}

if [[ ! $input == "y" ]]; then
  echo "Aborting."
  exit 0
else
  echo "Publishing message..."
  gcloud pubsub topics publish "$TOPIC" --message="$DATA" --project="$PROJECT_ID"

  update_history
  echo "Done."
fi
