package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"

	"github.com/Zomato/espresso/lib/workerpool"
	"github.com/Zomato/espresso/service/controller/pdf_generation"
	"github.com/Zomato/espresso/service/internal/pkg/viperpkg"
	"github.com/spf13/viper"
)

func main() {
	ctx := context.Background()

	viperpkg.InitConfig()

	log.Printf("Template storage type: %s", viper.GetString("template_storage.storage_type"))
	log.Printf("File storage type: %s", viper.GetString("file_storage.storage_type"))

	tabpool := viper.GetInt("browser.tab_pool")
	if err := browser_manager.Init(ctx, tabpool); err != nil {
		log.Fatalf("Failed to initialize browser: %v", err)
	}
	workerCount := viper.GetInt("workerpool.worker_count")
	workerTimeout := viper.GetInt("workerpool.worker_timeout")

	initializeWorkerPool(workerCount, workerTimeout)

	// register server for example v2
	// Create a new ServeMux
	mux := http.NewServeMux()

	pdf_generation.Register(mux)
	// Wrap the entire mux with the CORS middleware
	corsHandler := enableCORS(mux)

	log.Println("Starting PDF client server on :8081")
	if err := http.ListenAndServe(":8081", corsHandler); err != nil {
		log.Fatal(err)
	}

	// your implementation

	fmt.Println("Server terminated")
}

func initializeWorkerPool(workerCount int, workerTimeout int) {
	concurrency := workerCount

	workerpool.Initialize(concurrency,
		time.Duration(
			workerTimeout,
		)*time.Millisecond,
	)
}

// Create a global CORS middleware handler
func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		handler.ServeHTTP(w, r)
	})
}
