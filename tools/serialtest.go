package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"cdrgenerator/serial"
)

func main() {
	mode := flag.String("mode", "send", "Mode: send, receive, or loopback")
	device := flag.String("device", "/dev/ttyS0", "Serial device")
	baud := flag.Int("baud", 9600, "Baud rate")
	message := flag.String("message", "TEST", "Message to send")
	count := flag.Int("count", 10, "Number of messages")
	interval := flag.Duration("interval", 1*time.Second, "Interval between sends")
	flag.Parse()

	cfg := serial.PortConfig{
		Device:   *device,
		BaudRate: *baud,
		DataBits: 8,
		StopBits: 1,
		Parity:   "none",
	}

	switch *mode {
	case "send":
		sendTest(cfg, *message, *count, *interval)
	case "receive":
		receiveTest(cfg)
	case "loopback":
		loopbackTest(cfg, *message)
	default:
		log.Fatal("Invalid mode. Use: send, receive, or loopback")
	}
}

func sendTest(cfg serial.PortConfig, message string, count int, interval time.Duration) {
	port, err := serial.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open port: %v", err)
	}
	defer port.Close()

	fmt.Printf("Sending on %s at %d baud\n", cfg.Device, cfg.BaudRate)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Count: %d, Interval: %v\n\n", count, interval)

	for i := 0; i < count; i++ {
		msg := fmt.Sprintf("[%d] %s %s\n", i+1, message, time.Now().Format("15:04:05.000"))
		n, err := port.Write([]byte(msg))
		if err != nil {
			log.Printf("Write error: %v", err)
			continue
		}
		fmt.Printf("Sent %d bytes: %s", n, msg)
		time.Sleep(interval)
	}
	fmt.Println("\nSend test complete")
}

func receiveTest(cfg serial.PortConfig) {
	port, err := serial.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open port: %v", err)
	}
	defer port.Close()

	fmt.Printf("Listening on %s at %d baud\n", cfg.Device, cfg.BaudRate)
	fmt.Println("Press Ctrl+C to stop\n")

	buf := make([]byte, 1024)
	totalBytes := 0

	for {
		n, err := port.Read(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if n > 0 {
			totalBytes += n
			fmt.Printf("[%s] Received %d bytes (total: %d):\n", time.Now().Format("15:04:05.000"), n, totalBytes)
			fmt.Printf("%s\n", string(buf[:n]))
		}
	}
}

func loopbackTest(cfg serial.PortConfig, message string) {
	port, err := serial.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open port: %v", err)
	}
	defer port.Close()

	fmt.Printf("Loopback test on %s at %d baud\n", cfg.Device, cfg.BaudRate)
	fmt.Println("Connect pins 2 and 3 (TX and RX) with a jumper\n")

	for i := 0; i < 5; i++ {
		testMsg := fmt.Sprintf("%s-%d", message, i+1)
		fmt.Printf("Sending: %s\n", testMsg)

		// Send
		_, err := port.Write([]byte(testMsg + "\n"))
		if err != nil {
			log.Printf("Write error: %v", err)
			continue
		}

		// Try to receive with timeout
		time.Sleep(100 * time.Millisecond)
		buf := make([]byte, 256)
		n, err := port.Read(buf)
		if err != nil {
			fmt.Printf("  ✗ No data received (error: %v)\n", err)
		} else if n > 0 {
			received := string(buf[:n])
			if received == testMsg+"\n" {
				fmt.Printf("  ✓ Loopback OK: %s\n", received[:len(received)-1])
			} else {
				fmt.Printf("  ? Received different: %s\n", received)
			}
		} else {
			fmt.Printf("  ✗ No data received (timeout)\n")
		}

		time.Sleep(1 * time.Second)
	}
}
