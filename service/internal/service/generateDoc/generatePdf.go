package generateDoc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"
	"github.com/Zomato/espresso/lib/certmanager"
	"github.com/Zomato/espresso/lib/renderer"
	"github.com/Zomato/espresso/lib/signer"
	"github.com/Zomato/espresso/lib/templatestore"

	svcUtils "github.com/Zomato/espresso/service/utils"
	"github.com/go-rod/rod/lib/proto"
)

// GeneratePDF generates a PDF from the provided content and stores it in the provided file store.
// If signing is enabled, it will load the signing credentials in parallel and sign the PDF before storing it.
// The generated PDF is stored in the file store with the provided output file path.
// The function returns an error if anything goes wrong during generation, signing, or storage of the PDF.
func GeneratePDF(
	ctx context.Context,
	req *PDFDto,
	templateStoreAdapter *templatestore.StorageAdapter,
	fileStoreAdapter *templatestore.StorageAdapter,
	credentialStore *certmanager.CredentialStore,
) error {

	startTime := time.Now()

	// templateId := req.TemplateId
	content := req.Content
	viewPortConfig := req.ViewPort
	pdfParams := req.PdfParams
	viewPort := getViewPort(viewPortConfig)

	var credentials *certmanager.SigningCredentials
	var pdfReader io.Reader

	toBeSigned := req.SignParams != nil && req.SignParams.SignPdf
	if toBeSigned {
		credKey := req.SignParams.CertConfigKey
		var exists bool
		credentials, exists = credentialStore.GetCredential(credKey)
		if !exists {
			return fmt.Errorf("signing credentials not found for key: %s", credKey)
		}
	}

	var pdfSettings *proto.PagePrintToPDF
	if pdfParams != nil {
		pdfSettings = createPdfSettingsFromParams(pdfParams)
	} else {
		pdfSettings = &proto.PagePrintToPDF{}
	}

	pdfProps := renderer.GetHtmlPdfInput{
		TemplateRequest: templatestore.GetTemplateRequest{
			TemplatePath:   req.InputTemplatePath,
			TemplateS3Path: req.InputTemplatePath,
			TemplateBytes:  req.InputFileBytes,
			TemplateUUID:   req.InputTemplateUUID,
		},
		Data:         content,
		ViewPort:     viewPort,
		PdfParams:    pdfSettings,
		IsSinglePage: pdfParams.IsSinglePage,
	}

	pdf, err := renderer.GetHtmlPdf(ctx, &pdfProps, templateStoreAdapter)
	if err != nil {
		return fmt.Errorf("failed to generate pdf: %v", err)
	}
	defer pdf.Close()

	duration := time.Since(startTime)
	svcUtils.Logger.Info(ctx, "pdf stream received :: ", map[string]any{"duration": duration})

	duration = time.Since(startTime)

	if toBeSigned {
		signedPDF, err := signer.SignPdfStream(ctx, pdf, credentials.Certificate, credentials.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to sign pdf using SignPdfStream: %v", err)
		}

		pdfReader = bytes.NewReader(signedPDF)
	} else {
		pdfReader = pdf
	}
	svcUtils.Logger.Info(ctx, "starting upload :: ", map[string]any{"duration": duration})

	// Use the storage adapter to store the PDF
	docReq := &templatestore.PostDocumentRequest{
		FilePath:   req.OutputTemplatePath,
		FileS3Path: req.OutputTemplatePath,
	}
	// Upload the streaming data
	resp, err := (*fileStoreAdapter).PutDocument(ctx, docReq, &pdfReader)
	if err != nil {
		return fmt.Errorf("failed to store PDF: %v", err)
	}
	if resp == "stream" {
		req.OutputFileBytes = docReq.OutputFileBytes
	}

	duration = time.Since(startTime)
	svcUtils.Logger.Info(ctx, "uploaded to storage :: ", map[string]any{"duration": duration})

	return nil
}

func createPdfSettingsFromParams(pdfParams *PDFParams) *proto.PagePrintToPDF {

	pdfMarginTop := pdfParams.MarginTop
	pdfMarginBottom := pdfParams.MarginBottom
	pdfMarginLeft := pdfParams.MarginLeft
	pdfMarginRight := pdfParams.MarginRight
	pdfPaperWidth := pdfParams.PaperWidth
	pdfPaperHeight := pdfParams.PaperHeight

	pdfSettings := &proto.PagePrintToPDF{
		Landscape:           pdfParams.Landscape,
		DisplayHeaderFooter: pdfParams.DisplayHeaderFooter,
		PrintBackground:     pdfParams.PrintBackground,
		PageRanges:          pdfParams.PageRanges,
		HeaderTemplate:      pdfParams.HeaderTemplate,
		FooterTemplate:      pdfParams.FooterTemplate,
		PreferCSSPageSize:   pdfParams.PreferCssPageSize,
		MarginTop:           &pdfMarginTop,
		MarginBottom:        &pdfMarginBottom,
		MarginLeft:          &pdfMarginLeft,
		MarginRight:         &pdfMarginRight,
	}

	if pdfPaperWidth > 0 {
		pdfSettings.PaperWidth = &pdfPaperWidth
	}

	if pdfPaperHeight > 0 {
		pdfSettings.PaperHeight = &pdfPaperHeight
	}

	return pdfSettings
}

func getViewPort(viewPort *ViewportConfig) *browser_manager.ViewportConfig {

	viewSettings := &browser_manager.ViewportConfig{ // default viewport settings for A4 page
		Width:             794,
		Height:            1124,
		DeviceScaleFactor: 1.0,
		IsMobile:          false,
	}

	if viewPort == nil {
		return viewSettings
	}

	viewSettings = &browser_manager.ViewportConfig{
		Width:             int(viewPort.Width),
		Height:            int(viewPort.Height),
		DeviceScaleFactor: viewPort.DeviceScaleFactor,
		IsMobile:          viewPort.IsMobile,
	}

	return viewSettings
}

func SignPDF(
	ctx context.Context,
	req *SignPDFDto,
	fileStoreAdapter *templatestore.StorageAdapter,
	credentialStore *certmanager.CredentialStore,
) error {

	reqId := req.ReqId
	svcUtils.Logger.Info(ctx, "SignPDF called ", map[string]any{"req id": reqId})

	// get input file stream
	freader, err := (*fileStoreAdapter).GetDocument(ctx, &templatestore.GetDocumentRequest{
		FilePath:       req.InputFilePath,
		FileS3Path:     req.InputFilePath,
		InputFileBytes: req.InputFileBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to get input file: %v", err)
	}

	var credentials *certmanager.SigningCredentials
	var pdfReader io.Reader

	if req.SignParams.SignPdf {
		credKey := req.SignParams.CertConfigKey
		var exists bool
		credentials, exists = credentialStore.GetCredential(credKey)
		if !exists {
			return fmt.Errorf("signing credentials not found for key: %s", credKey)
		}

		// convert pdfreader to *rod.StreamReader
		signedPDF, err := signer.SignPdfStream(ctx, freader, credentials.Certificate, credentials.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to sign pdf using SignPdfStream: %v", err)
		}

		pdfReader = bytes.NewReader(signedPDF)
	} else {
		pdfReader = freader
	}
	// Use the storage adapter to store the PDF
	docReq := &templatestore.PostDocumentRequest{
		FilePath:   req.OutputFilePath,
		FileS3Path: req.OutputFilePath,
	}

	// Upload the streaming data
	resp, err := (*fileStoreAdapter).PutDocument(ctx, docReq, &pdfReader)
	if err != nil {
		return fmt.Errorf("failed to store PDF: %v", err)
	}
	if resp == "stream" {
		req.OutputFileBytes = docReq.OutputFileBytes
	}

	return nil
}
