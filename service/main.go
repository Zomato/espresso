package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"

	logger "github.com/Zomato/espresso/lib/logger"
	"github.com/Zomato/espresso/lib/workerpool"
	"github.com/Zomato/espresso/service/pkg/config"
	"github.com/Zomato/espresso/service/server"
	"github.com/Zomato/espresso/service/utils"
)

func main() {
	ctx := context.Background()

	config.InitConfig()
	cfg := config.GetConfig()
	// Replace ZeroLog with any logging library by implementing ILogger interface.
	zeroLog := utils.NewZeroLogger()
	logger.Initialize(zeroLog)

	templateStorageType := cfg.TemplateStorage.StorageType
	zeroLog.Info(ctx, "Template storage type ", map[string]any{"type": templateStorageType})

	fileStorageType := cfg.FileStorage.StorageType
	zeroLog.Info(ctx, "File storage type ", map[string]any{"type": fileStorageType})

	tabpool := cfg.Browser.TabPool
	if err := browser_manager.Init(ctx, tabpool); err != nil {
		log.Fatalf("Failed to initialize browser: %v", err)
	}
	workerCount := cfg.WorkerPool.WorkerCount
	workerTimeout := cfg.WorkerPool.WorkerTimeout

	initializeWorkerPool(workerCount, workerTimeout)

	// register server for example v2
	// Create a new ServeMux
	mux := http.NewServeMux()

	server.RegisterHTTP(mux)
	if cfg.MCP.Enabled {
		zeroLog.Info(ctx, "MCP is enabled. Initializing MCP components...", nil)
		server.RegisterMCP(mux)
	}
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
