package generateDoc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"
	"github.com/Zomato/espresso/lib/certmanager"
	"github.com/Zomato/espresso/lib/renderer"
	"github.com/Zomato/espresso/lib/signer"
	"github.com/Zomato/espresso/lib/templatestore"
	"github.com/Zomato/espresso/lib/workerpool"
	"github.com/spf13/viper"

	"github.com/go-rod/rod/lib/proto"
)

// GeneratePDF generates a PDF from the provided content and stores it in the provided file store.
// If signing is enabled, it will load the signing credentials in parallel and sign the PDF before storing it.
// The generated PDF is stored in the file store with the provided output file path.
// The function returns an error if anything goes wrong during generation, signing, or storage of the PDF.
func GeneratePDF(ctx context.Context, req *PDFDto, templateStoreAdapter *templatestore.StorageAdapter, fileStoreAdapter *templatestore.StorageAdapter) error {

	startTime := time.Now()

	// templateId := req.TemplateId
	content := req.Content
	viewPortConfig := req.ViewPort
	pdfParams := req.PdfParams

	viewPort := getViewPort(viewPortConfig)
	// Start loading credentials in parallel if signing is enabled
	var credWg sync.WaitGroup
	var credErr error
	var credentials *certmanager.SigningCredentials
	var pdfReader io.Reader
	toBeSigned := false

	if req.SignParams != nil && req.SignParams.SignPdf {
		toBeSigned = true
	}
	if toBeSigned {
		certConfig := &certmanager.CertificateConfig{
			CertFilePath: viper.GetString(req.SignParams.CertConfigKey + ".cert_filepath"),
			KeyFilePath:  viper.GetString(req.SignParams.CertConfigKey + ".key_filepath"),
			KeyPassword:  viper.GetString(req.SignParams.CertConfigKey + ".key_password"),
		}
		credWg.Add(1)
		err := workerpool.Pool().SubmitTask(
			func(args ...interface{}) {
				defer credWg.Done()
				ctxArg := args[0].(context.Context)
				credentials, credErr = certmanager.LoadSigningCredentials(ctxArg, certConfig)
			},
			ctx,
		)

		if err != nil {
			return fmt.Errorf("failed to submit credential loading task: %v", err)
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
	fmt.Println("pdf stream received at :: ", duration)

	duration = time.Since(startTime)

	if toBeSigned {
		credWg.Wait()

		if credErr != nil {
			return fmt.Errorf("failed to load signing credentials: %v", credErr)
		}

		signedPDF, err := signer.SignPdfStream(ctx, pdf, credentials.Certificate, credentials.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to sign pdf using SignPdfStream: %v", err)
		}

		pdfReader = bytes.NewReader(signedPDF)
	} else {
		pdfReader = pdf
	}
	fmt.Println("starting upload :: ", duration)
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
	fmt.Println("uploaded to storage at :: ", duration)

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

func SignPDF(ctx context.Context, req *SignPDFDto, fileStoreAdapter *templatestore.StorageAdapter) error {

	reqId := req.ReqId
	fmt.Println("SignPDF called, req id :: ", reqId)
	// get input file stream
	freader, err := (*fileStoreAdapter).GetDocument(ctx, &templatestore.GetDocumentRequest{
		FilePath:       req.InputFilePath,
		FileS3Path:     req.InputFilePath,
		InputFileBytes: req.InputFileBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to get input file: %v", err)
	}
	// Start loading credentials in parallel if signing is enabled
	var credWg sync.WaitGroup
	var credErr error
	var credentials *certmanager.SigningCredentials
	var pdfReader io.Reader

	if req.SignParams.SignPdf {
		credWg.Add(1)
		certConfig := &certmanager.CertificateConfig{
			CertFilePath: viper.GetString(req.SignParams.CertConfigKey + ".cert_filepath"),
			KeyFilePath:  viper.GetString(req.SignParams.CertConfigKey + ".key_filepath"),
			KeyPassword:  viper.GetString(req.SignParams.CertConfigKey + ".key_password"),
		}
		err := workerpool.Pool().SubmitTask(
			func(args ...interface{}) {
				defer credWg.Done()
				ctxArg := args[0].(context.Context)
				credentials, credErr = certmanager.LoadSigningCredentials(ctxArg, certConfig)
			},
			ctx,
		)

		if err != nil {
			return fmt.Errorf("failed to submit credential loading task: %v", err)
		}
	}
	if req.SignParams.SignPdf {
		credWg.Wait()

		if credErr != nil {
			return fmt.Errorf("failed to load signing credentials: %v", credErr)
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
