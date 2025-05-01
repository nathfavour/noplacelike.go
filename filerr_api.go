package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Pairing endpoints
func InitiatePairing(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"qr_code": "mock-qr-code-string",
		"message": "Scan this QR code to pair device.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func CompletePairing(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":  "success",
		"message": "Pairing completed successfully.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// File operations
func ListFiles(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"files": []map[string]interface{}{
			{"name": "file1.txt", "type": "file", "size": 1234},
			{"name": "folder1", "type": "folder"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func SendFiles(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"transfer_id": "mock-transfer-id-123",
		"status":      "started",
		"message":     "File transfer started.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ReceiveFiles(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"transfer_id": "mock-transfer-id-456",
		"status":      "started",
		"message":     "Ready to receive files.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Transfer status
func GetTransferStatus(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"transfer_id": "mock-transfer-id-123",
		"progress":    42,
		"status":      "in_progress",
		"speed":       "2MB/s",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Directory monitoring
func StartMonitoring(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":  "monitoring_started",
		"message": "Directory monitoring started.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetMonitorStatus(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"changes": []map[string]interface{}{
			{"name": "file2.txt", "event": "created"},
			{"name": "file1.txt", "event": "modified"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// WebSocket for real-time updates
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TransferWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for i := 0; i <= 100; i += 10 {
		msg := map[string]interface{}{
			"transfer_id": "mock-transfer-id-123",
			"progress":    i,
			"status":      "in_progress",
		}
		conn.WriteJSON(msg)
		time.Sleep(300 * time.Millisecond)
	}
	conn.WriteJSON(map[string]interface{}{
		"transfer_id": "mock-transfer-id-123",
		"progress":    100,
		"status":      "completed",
	})
}

func RegisterFilerrAPI(router *mux.Router) {
	router.HandleFunc("/pair/initiate", InitiatePairing).Methods("POST")
	router.HandleFunc("/pair/complete", CompletePairing).Methods("POST")
	router.HandleFunc("/files/list", ListFiles).Methods("GET")
	router.HandleFunc("/files/send", SendFiles).Methods("POST")
	router.HandleFunc("/files/receive", ReceiveFiles).Methods("POST")
	router.HandleFunc("/transfer/status/{id}", GetTransferStatus).Methods("GET")
	router.HandleFunc("/monitor/start", StartMonitoring).Methods("POST")
	router.HandleFunc("/monitor/status", GetMonitorStatus).Methods("GET")
	router.HandleFunc("/ws/transfer", TransferWebSocket)
}
