package renderer

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func ExecuteTemplate(ctx context.Context, templateFile *template.Template, data map[string]interface{}) (string, error) {

	// Validate template and data
	if templateFile == nil {
		return "", fmt.Errorf("template file is nil")
	}

	// Get buffer from pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Execute template with buffered writer
	if err := templateFile.Execute(buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func AddImagesFromMetaData(ctx context.Context, htmlContent string, unmarshaledData map[string]interface{}) string {

	if metadata, ok := unmarshaledData["metadata"].(map[string]interface{}); ok {
		if images, ok := metadata["images"].(map[string]interface{}); ok {
			for url, dataURI := range images {
				if dataURIStr, ok := dataURI.(string); ok {
					htmlContent = strings.Replace(htmlContent, url, dataURIStr, -1)
				}
			}
		}
	}
	return htmlContent
}
