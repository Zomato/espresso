package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"

	logger "github.com/Zomato/espresso/lib/logger"
	"github.com/Zomato/espresso/lib/workerpool"
	"github.com/Zomato/espresso/service/controller/pdf_generation"
	"github.com/Zomato/espresso/service/internal/config"
	"github.com/Zomato/espresso/service/utils"
)

func main() {
	ctx := context.Background()

	config, err := config.Load("/app/espresso/configs")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Replace ZeroLog with any logging library by implementing ILogger interface.
	zeroLog := utils.NewZeroLogger(config.AppConfig.LogLevel)
	logger.Initialize(zeroLog)

	log.Printf("Template storage type: %s", config.TemplateStorageConfig.StorageType)
	log.Printf("File storage type: %s", config.FileStorageConfig.StorageType)

	tabpool := config.BrowserConfig.TabPool

	if err := browser_manager.Init(ctx, tabpool); err != nil {
		log.Fatalf("Failed to initialize browser: %v", err)
	}

	workerCount := config.WorkerPoolConfig.WorkerCount
	workerTimeout := config.WorkerPoolConfig.WorkerTimeoutMs

	initializeWorkerPool(workerCount, workerTimeout)

	// register server for example v2
	// Create a new ServeMux
	mux := http.NewServeMux()

	pdf_generation.Register(mux, config)
	// Wrap the entire mux with the CORS middleware
	corsHandler := enableCORS(mux)

	log.Println("Starting PDF client server on :8081")
	if err := http.ListenAndServe(":8081", corsHandler); err != nil {
		log.Fatal(err)
	}

	// your implementation

	zeroLog.Info(ctx, "Server terminated", nil)
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
