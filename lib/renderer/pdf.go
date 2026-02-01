package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"
	log "github.com/Zomato/espresso/lib/logger"
	"github.com/Zomato/espresso/lib/templatestore"
)

func GetHtmlPdf(ctx context.Context, params *GetHtmlPdfInput, storeAdapter *templatestore.StorageAdapter) ([]byte, error) {

	startTime := time.Now()
	if params == nil {
		return nil, fmt.Errorf("params are required")
	}

	duration := time.Since(startTime)

	log.Logger.Info(ctx, "starting template parsing at", map[string]any{"duration": duration})

	var err error
	var templateFile *template.Template
	if storeAdapter != nil {
		templateFile, err = (*storeAdapter).GetTemplate(ctx, &params.TemplateRequest)
		if err != nil {
			return nil, fmt.Errorf("unable to get template file from store: %v", err)
		}
	} else {
		if len(params.TemplateRequest.TemplateBytes) > 0 {
			templateFile, err = template.New("stream").Parse(string(params.TemplateRequest.TemplateBytes))
			if err != nil {
				return nil, fmt.Errorf("unable to parse template file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("storage configuration is invalid")
		}
	}

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "starting unmarshaling data at", map[string]any{"duration": duration})

	data := params.Data

	var unmarshaledData map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaledData); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON data: %v", err)
	}

	metaInfo := getMetaInfo(unmarshaledData)
	if metaInfo != nil {
		unmarshaledData["metadata"] = metaInfo
	}

	page := browser_manager.GetTab()
	defer func() {
		duration = time.Since(startTime)
		log.Logger.Info(ctx, "closing tab at", map[string]any{"duration": duration})
		browser_manager.ReleaseTab(page)
	}()

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "prefetching images at", map[string]any{"duration": duration})
	unmarshaledData = PrefetchImages(ctx, unmarshaledData)

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "unmarshaled data & started template execution at", map[string]any{"duration": duration})

	htmlContent, err := ExecuteTemplate(ctx, templateFile, unmarshaledData)
	if err != nil {
		return nil, fmt.Errorf("unable to execute template file: %v", err)
	}

	htmlContent = AddImagesFromMetaData(ctx, htmlContent, unmarshaledData)

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "template executed and requesting new tab at", map[string]any{"duration": duration})

	if params.IsSinglePage {
		page.MustSetViewport(794, 1124, 1.0, false)
	} else {
		viewPortConfig := params.ViewPort
		page.MustSetViewport(viewPortConfig.Width, viewPortConfig.Height, viewPortConfig.DeviceScaleFactor, viewPortConfig.IsMobile)
	}

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "rendering data in new tab at", map[string]any{"duration": duration})

	err = page.SetDocumentContent(string(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("unable to generate pdf: %v", err)
	}

	pdfParams := params.PdfParams

	if params.IsSinglePage { // to generate pdf of single page with dynamic height

		err = page.WaitLoad()
		if err != nil {
			return nil, fmt.Errorf("error in waiting for page load: %v", err)
		}

		body, err := page.Element("html")
		if err != nil {
			return nil, fmt.Errorf("error in getting html element: %v", err)
		}

		heightProp, err := body.Property("scrollHeight")
		if err != nil {
			return nil, fmt.Errorf("error in getting scroll height: %v", err)
		}

		pdfHeight := heightProp.Num()

		dynamicHeight := float64(pdfHeight / 96)
		pdfParams.PaperHeight = &dynamicHeight

	}

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "generating pdf at", map[string]any{"duration": duration})

	pdfStream, err := page.PDF(pdfParams)
	if err != nil {
		return nil, fmt.Errorf("unable to generate pdf: %v", err)
	}

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "reading pdf stream at", map[string]any{"duration": duration})

	// Read the stream fully BEFORE releasing the tab to prevent memory leak
	pdfBytes, err := io.ReadAll(pdfStream)
	if err != nil {
		return nil, fmt.Errorf("unable to read pdf stream: %v", err)
	}

	// Close the stream while the page is still alive to properly cleanup Chrome's IO handle
	if closeErr := pdfStream.Close(); closeErr != nil {
		log.Logger.Error(ctx, "failed to close pdf stream", closeErr, nil)
	}

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "pdf generated at", map[string]any{"duration": duration})

	return pdfBytes, nil
}

func getMetaInfo(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	if data["metadata"] == nil {
		return nil
	}

	metaInfo := make(map[string]interface{})

	if metaData, ok := data["metadata"]; ok {
		metaDataMap, ok := metaData.(map[string]interface{})
		if !ok {
			return nil
		}

		for key, value := range metaDataMap {
			if key == "images" {
				if images, ok := value.([]interface{}); ok {
					imageMap := make(map[string]interface{})
					for _, img := range images {
						if url, ok := img.(string); ok {
							imageMap[url] = url
						}
					}
					metaInfo[key] = imageMap
				}
			} else {
				metaInfo[key] = value
			}
		}
	}

	return metaInfo
}
