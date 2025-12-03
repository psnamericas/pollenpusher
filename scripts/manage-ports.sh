#!/bin/bash

# CDR Generator Port Management Script
# Usage: ./manage-ports.sh [list|add-replay|add-synthetic|remove]

API_URL="http://100.88.226.119:8080/api/config"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function list_ports() {
    echo -e "${BLUE}Current Port Configuration:${NC}"
    curl -s $API_URL | jq -r '.ports[] | "\(.device) - \(.format) (\(.mode)) - \(if .enabled then "ENABLED" else "DISABLED" end)"'
}

function add_replay_port() {
    echo -e "${GREEN}Adding Replay Port${NC}"
    read -p "Device (e.g., /dev/ttyS5): " device
    read -p "Format (vesta/viper): " format
    read -p "Baud Rate (default: 9600): " baud_rate
    baud_rate=${baud_rate:-9600}
    read -p "Sample File Path: " sample_file
    read -p "Calls Per Minute (default: 2.5): " cpm
    cpm=${cpm:-2.5}
    read -p "Description: " description

    # Get current config
    config=$(curl -s $API_URL)

    # Add new port
    new_port=$(cat <<EOF
{
  "device": "$device",
  "baud_rate": $baud_rate,
  "format": "$format",
  "mode": "replay",
  "sample_file": "$sample_file",
  "loop": true,
  "calls_per_minute": $cpm,
  "enabled": true,
  "description": "$description"
}
EOF
)

    # Merge and save
    updated_config=$(echo "$config" | jq ".ports += [$new_port]")

    echo "$updated_config" | curl -s -X POST -H "Content-Type: application/json" -d @- $API_URL | jq .

    echo -e "${YELLOW}Configuration saved! Restart the service to apply changes:${NC}"
    echo "  ssh root@100.88.226.119 'systemctl restart cdrgenerator.service'"
}

function add_synthetic_port() {
    echo -e "${GREEN}Adding Synthetic Port${NC}"
    read -p "Device (e.g., /dev/ttyS5): " device
    read -p "Format (vesta/viper): " format
    read -p "Baud Rate (default: 9600): " baud_rate
    baud_rate=${baud_rate:-9600}
    read -p "System ID: " system_id
    read -p "Agent Count (default: 10): " agent_count
    agent_count=${agent_count:-10}
    read -p "Min Duration (sec, default: 10): " min_dur
    min_dur=${min_dur:-10}
    read -p "Max Duration (sec, default: 300): " max_dur
    max_dur=${max_dur:-300}
    read -p "Calls Per Minute (default: 2.5): " cpm
    cpm=${cpm:-2.5}
    read -p "Description: " description

    # Get current config
    config=$(curl -s $API_URL)

    # Add new port
    new_port=$(cat <<EOF
{
  "device": "$device",
  "baud_rate": $baud_rate,
  "format": "$format",
  "mode": "synthetic",
  "calls_per_minute": $cpm,
  "enabled": true,
  "description": "$description",
  "synthetic": {
    "system_id": "$system_id",
    "agent_count": $agent_count,
    "min_duration_sec": $min_dur,
    "max_duration_sec": $max_dur,
    "include_agent_events": true
  }
}
EOF
)

    # Merge and save
    updated_config=$(echo "$config" | jq ".ports += [$new_port]")

    echo "$updated_config" | curl -s -X POST -H "Content-Type: application/json" -d @- $API_URL | jq .

    echo -e "${YELLOW}Configuration saved! Restart the service to apply changes:${NC}"
    echo "  ssh root@100.88.226.119 'systemctl restart cdrgenerator.service'"
}

function remove_port() {
    echo -e "${BLUE}Current Ports:${NC}"
    curl -s $API_URL | jq -r '.ports | to_entries[] | "\(.key): \(.value.device) - \(.value.format)"'

    read -p "Enter port index to remove: " index

    # Get current config
    config=$(curl -s $API_URL)

    # Remove port
    updated_config=$(echo "$config" | jq "del(.ports[$index])")

    echo "$updated_config" | curl -s -X POST -H "Content-Type: application/json" -d @- $API_URL | jq .

    echo -e "${GREEN}Port removed! Restart the service to apply changes.${NC}"
}

# Main menu
case "$1" in
    list)
        list_ports
        ;;
    add-replay)
        add_replay_port
        ;;
    add-synthetic)
        add_synthetic_port
        ;;
    remove)
        remove_port
        ;;
    *)
        echo "Usage: $0 {list|add-replay|add-synthetic|remove}"
        echo ""
        echo "Examples:"
        echo "  $0 list                  - List all configured ports"
        echo "  $0 add-replay            - Add a new port with sample data (replay mode)"
        echo "  $0 add-synthetic         - Add a new port with synthetic data"
        echo "  $0 remove                - Remove a port"
        exit 1
        ;;
esac
