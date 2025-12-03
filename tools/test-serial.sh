#!/bin/bash

# Serial Port Testing Script
# Usage: ./test-serial.sh <mode> <device>
# Modes: send, receive, info

MODE=$1
DEVICE=$2

if [ -z "$MODE" ] || [ -z "$DEVICE" ]; then
    echo "Usage: $0 <send|receive|info> <device>"
    echo "Example: $0 send /dev/ttyS3"
    echo "Example: $0 receive /dev/ttyS4"
    echo "Example: $0 info /dev/ttyS3"
    exit 1
fi

# Check if device exists
if [ ! -e "$DEVICE" ]; then
    echo "Error: Device $DEVICE does not exist"
    exit 1
fi

case "$MODE" in
    send)
        echo "=== Serial Port SEND Test ==="
        echo "Device: $DEVICE"
        echo "Configuring port: 9600 baud, 8N1"
        stty -F $DEVICE 9600 cs8 -cstopb -parenb clocal -crtscts

        echo "Sending 10 test messages (1 per second)..."
        for i in {1..10}; do
            MSG="[TEST-$i] $(date '+%Y-%m-%d %H:%M:%S.%N')"
            echo "$MSG" > $DEVICE
            echo "Sent: $MSG"
            sleep 1
        done
        echo "Send test complete"
        ;;

    receive)
        echo "=== Serial Port RECEIVE Test ==="
        echo "Device: $DEVICE"
        echo "Configuring port: 9600 baud, 8N1"
        stty -F $DEVICE 9600 cs8 -cstopb -parenb clocal -crtscts

        echo "Listening for data... (Press Ctrl+C to stop)"
        echo "Start time: $(date '+%H:%M:%S')"
        echo ""

        # Use timeout and show hex + ASCII
        timeout 30 cat $DEVICE | while IFS= read -r line; do
            echo "[$(date '+%H:%M:%S')] $line"
        done

        if [ ${PIPESTATUS[0]} -eq 124 ]; then
            echo ""
            echo "Timeout after 30 seconds - no data received"
        fi
        ;;

    info)
        echo "=== Serial Port INFO ==="
        echo "Device: $DEVICE"
        echo ""

        # Device file info
        echo "File permissions:"
        ls -l $DEVICE
        echo ""

        # Port settings
        echo "Current port settings:"
        stty -F $DEVICE -a
        echo ""

        # Kernel stats
        DEVICE_NUM=$(echo $DEVICE | grep -o '[0-9]\+$')
        echo "Kernel driver statistics:"
        grep "^[[:space:]]*${DEVICE_NUM}:" /proc/tty/driver/serial
        echo ""

        # dmesg info
        echo "Kernel messages for this port:"
        dmesg | grep "ttyS${DEVICE_NUM}" | tail -5
        ;;

    *)
        echo "Error: Invalid mode '$MODE'"
        echo "Valid modes: send, receive, info"
        exit 1
        ;;
esac
