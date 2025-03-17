package renderer

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"
	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/Zomato/espresso/lib/workerpool"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
)

func TestGetHtmlPdf(t *testing.T) {
	ctx := context.Background()
	os.Setenv("ROD_BROWSER_BIN", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	err := browser_manager.Init(ctx, 1)
	assert.NoError(t, err)
	concurrency := 2

	workerpool.Initialize(concurrency,
		time.Duration(
			200,
		)*time.Millisecond,
	)

	tests := []struct {
		name        string
		input       *GetHtmlPdfInput
		wantErr     bool
		description string
	}{
		{
			name: "basic_template",
			input: &GetHtmlPdfInput{
				TemplateRequest: templatestore.GetTemplateRequest{
					TemplateBytes: []byte(`<html><body><h1>{{.title}}</h1></body></html>`),
				},
				Data: []byte(`{"title":"Test Document"}`),
				ViewPort: &browser_manager.ViewportConfig{
					Width:             794,
					Height:            1124,
					DeviceScaleFactor: 1.0,
				},
				PdfParams: &proto.PagePrintToPDF{
					PrintBackground: true,
					MarginTop:       float64Ptr(0.4),
					MarginBottom:    float64Ptr(0.4),
				},
			},
			wantErr:     false,
			description: "Should generate PDF from basic template",
		},
		{
			name: "invalid_template",
			input: &GetHtmlPdfInput{
				TemplateRequest: templatestore.GetTemplateRequest{
					TemplateBytes: []byte(`<html><body><h1>{{.title}</h1></body></html>`), // Invalid template syntax - missing closing brace
				},
				Data: []byte(`{"title":"Test Document"}`),
				ViewPort: &browser_manager.ViewportConfig{
					Width:             794,
					Height:            1124,
					DeviceScaleFactor: 1.0,
				},
				PdfParams: &proto.PagePrintToPDF{
					PrintBackground: true,
					MarginTop:       float64Ptr(0.4),
					MarginBottom:    float64Ptr(0.4),
				},
			},
			wantErr:     true,
			description: "Should fail with invalid template syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdf, err := GetHtmlPdf(ctx, tt.input, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, pdf)
			defer pdf.Close()

			// Read first few bytes to verify it's a PDF
			buf := make([]byte, 4)
			_, err = pdf.Read(buf)
			assert.NoError(t, err)
			assert.Equal(t, []byte("%PDF"), buf)
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}
