package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

var activeConn *websocket.Conn
var pendingRequests = make(map[string]chan Response)
var mu sync.Mutex

// Instruction represents a command sent to the agent
type Instruction struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Response represents a message received from the agent
type Response struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("Agent connected")
	activeConn = conn

	// Listen for responses
	for {
		var resp Response
		err := conn.ReadJSON(&resp)
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Agent response: Status=%s, Message=%s\n", resp.Status, resp.Message)

		if resp.ID != "" {
			mu.Lock()
			if ch, ok := pendingRequests[resp.ID]; ok {
				ch <- resp
				delete(pendingRequests, resp.ID)
			}
			mu.Unlock()
		}
	}
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if activeConn == nil {
		http.Error(w, "No agent connected", http.StatusServiceUnavailable)
		return
	}

	var instruction Instruction
	if err := json.NewDecoder(r.Body).Decode(&instruction); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	instruction.ID = uuid.New().String()

	respCh := make(chan Response)
	mu.Lock()
	pendingRequests[instruction.ID] = respCh
	mu.Unlock()

	if err := activeConn.WriteJSON(instruction); err != nil {
		log.Println("Write error:", err)
		mu.Lock()
		delete(pendingRequests, instruction.ID)
		mu.Unlock()
		http.Error(w, "Failed to send instruction", http.StatusInternalServerError)
		return
	}

	select {
	case resp := <-respCh:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case <-time.After(10 * time.Second):
		mu.Lock()
		delete(pendingRequests, instruction.ID)
		mu.Unlock()
		http.Error(w, "Timeout waiting for agent response", http.StatusGatewayTimeout)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/execute", handleExecute)
	port := ":8080"
	log.Printf("Controller listening on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
