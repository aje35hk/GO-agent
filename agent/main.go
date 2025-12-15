package main

import (
	"agent/pkg/k8s"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"os/exec"

	"github.com/gorilla/websocket"
)

// Instruction represents a command received from the controller
type Instruction struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Response represents a message sent to the controller
type Response struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func main() {
	// Initialize Kubernetes client
	// NOTE: This requires a 'k8s' package with an InitClient function.
	// For example: import k8s "github.com/your-org/your-repo/pkg/k8s"
	// If you don't have this package, you'll need to create it or remove this block.
	_, err := k8s.InitClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize Kubernetes client: %v", err)
		// We continue even if K8s init fails, as we might want to run other commands or retry later
	}

	controllerAddr := flag.String("controller", "commander-service:8080", "Controller address")
	flag.Parse()

	u := url.URL{
		Scheme: "ws",
		Host:   *controllerAddr,
		Path:   "/ws",
	}

	log.Printf("Connecting to %s", u.String())

	// Retry connection logic
	var c *websocket.Conn
	// err is already declared above for k8s.InitClient, so we reuse it.
	for {
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			break
		}
		log.Printf("Connection failed: %v. Retrying in 2 seconds...", err)
		time.Sleep(2 * time.Second)
	}
	defer c.Close()

	log.Println("Connected to controller")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var instruction Instruction
			err := c.ReadJSON(&instruction)
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			log.Printf("Received instruction: %s %s\n", instruction.Type, instruction.Payload)

			var output []byte
			var status, message string
			status = "success"
			message = "Command executed"

			if instruction.Type == "print" && instruction.Payload == "hello" {
				fmt.Println("Hello World")
				message = "Done"
			} else if instruction.Type == "bash" {
				cmd := exec.Command("bash", "-c", instruction.Payload)
				output, err = cmd.CombinedOutput()
				if err != nil {
					status = "error"
					message = err.Error()
				}
			} else if instruction.Type == "kubectl" {
				// Execute kubectl command
				// We assume payload contains arguments like "get pods", "logs my-pod", etc.
				// Security Warning: detailed input validation should be implemented in production
				cmd := exec.Command("bash", "-c", "kubectl "+instruction.Payload)
				output, err = cmd.CombinedOutput()
				if err != nil {
					status = "error"
					message = err.Error()
				}
			} else {
				status = "error"
				message = "Unknown instruction type"
			}

			resp := Response{
				ID:      instruction.ID,
				Status:  status,
				Message: message,
				Output:  string(output),
			}
			if err := c.WriteJSON(resp); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-done:
	case <-interrupt:
		log.Println("Interrupt received, closing connection...")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Write close error:", err)
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}
