package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Job struct {
	SellerAsset string `json:"seller_asset"`
	BuyerUpdate string `json:"buyer_update"`
	OrderID     string `json:"order_id"`
	HandoffOK   bool   `json:"handoff_ok"`
}

type Decision struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func decide(j Job) Decision {
	if j.SellerAsset == "" || j.BuyerUpdate == "" || j.OrderID == "" {
		return Decision{Status: "rejected", Reason: "seller asset, buyer update, and order id are required"}
	}
	if !j.HandoffOK {
		return Decision{Status: "failed", Reason: "order handoff was not confirmed"}
	}
	return Decision{Status: "completed"}
}

func runHandler(c *InfraiClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var job Job
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			http.Error(w, "invalid job", http.StatusBadRequest)
			return
		}
		result := decide(job)
		if result.Status == "failed" {
			if err := c.Capture(job, result.Reason); err != nil {
				log.Printf("infrai capture failed: %v", err)
				http.Error(w, "error capture failed", http.StatusBadGateway)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	c := &InfraiClient{Key: key, HTTP: &http.Client{Timeout: 10 * time.Second}}
	log.Println("marketplace job service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", runHandler(c)))
}
