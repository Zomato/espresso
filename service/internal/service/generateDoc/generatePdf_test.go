package generateDoc

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"
	log "github.com/Zomato/espresso/lib/logger"
	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/Zomato/espresso/lib/workerpool"
	svcUtils "github.com/Zomato/espresso/service/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func float64Ptr(v float64) *float64 {
	return &v
}

func TestGeneratePDF_LoggingOutputs(t *testing.T) {
	ctx := context.Background()

	// Initialize browser and worker pool
	os.Setenv("ROD_BROWSER_BIN", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	err := browser_manager.Init(ctx, 1)
	require.NoError(t, err)

	workerpool.Initialize(2, 200*time.Millisecond)

	// Set up memory buffer for intercepting ZeroLog output
	var logBuf bytes.Buffer
	logger := svcUtils.NewZeroLoggerWithConfig(svcUtils.LogConfig{
		Disabled: false,
		Level:    "info",
		Format:   "json",
		Output:   &logBuf,
	})
	log.Initialize(logger)

	fileStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{
		StorageType: "stream",
	})
	require.NoError(t, err)

	req := &PDFDto{
		ReqId: "test-feature-log-req",
		Content: []byte(`{"title": "Logging Test"}`),
		InputFileBytes: []byte(`<html><body><h1>{{.title}}</h1></body></html>`),
		PdfParams: &PDFParams{
			PrintBackground: true,
			MarginTop:       0.4,
			MarginBottom:    0.4,
		},
	}

	err = GeneratePDF(ctx, req, nil, &fileStorageAdapter)
	require.NoError(t, err)
	assert.NotEmpty(t, req.OutputFileBytes)

	// Assert on intercepted logs
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "pdf bytes received :: ")
	assert.Contains(t, logOutput, "starting upload :: ")
	assert.Contains(t, logOutput, "uploaded to storage :: ")

	// Test that when disabled, no logs are written during PDF generation
	logBuf.Reset()
	disabledLogger := svcUtils.NewZeroLoggerWithConfig(svcUtils.LogConfig{
		Disabled: true,
		Output:   &logBuf,
	})
	log.Initialize(disabledLogger)

	err = GeneratePDF(ctx, req, nil, &fileStorageAdapter)
	require.NoError(t, err)
	assert.Empty(t, logBuf.String(), "Expected no logs when logger is disabled")
}
