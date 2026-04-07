package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Zomato/espresso/lib/templatestore"
	libutils "github.com/Zomato/espresso/lib/utils"
	"github.com/Zomato/espresso/service/dto"
	"github.com/Zomato/espresso/service/internal/service/generateDoc"
	svcUtils "github.com/Zomato/espresso/service/utils"
)

type PDFService struct {
	TemplateStorageAdapter *templatestore.StorageAdapter
	FileStorageAdapter     *templatestore.StorageAdapter
}

func NewPDFService(templateAdapter, fileAdapter templatestore.StorageAdapter) *PDFService {
	return &PDFService{TemplateStorageAdapter: &templateAdapter, FileStorageAdapter: &fileAdapter}
}

func (s *PDFService) GeneratePDF(ctx context.Context, req *dto.GeneratePDFRequest) (*dto.GeneratePDFResponse, error) {
	startTime := time.Now()

	reqID := libutils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "GeneratePDF called :: ", map[string]any{"req_id": reqID})

	generatePDFReq := &generateDoc.PDFDto{
		ReqId:              reqID,
		InputTemplatePath:  req.InputFilePath,
		InputFileBytes:     req.InputFileBytes,
		InputTemplateUUID:  req.InputTemplateUuid,
		OutputTemplatePath: req.OutputFilePath,
		Content:            req.Content,
		ViewPort:           req.Viewport,
		PdfParams:          req.PdfParams,
	}

	if req.SignParams != nil && req.SignParams.SignPdf {
		generatePDFReq.SignParams = req.SignParams
	}

	if err := generateDoc.GeneratePDF(ctx, generatePDFReq, s.TemplateStorageAdapter, s.FileStorageAdapter); err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf", err, nil)
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	duration := time.Since(startTime)
	svcUtils.Logger.Info(ctx, "generated pdf :: ", map[string]any{"req_id": reqID, "duration": duration})
	return &dto.GeneratePDFResponse{
		Status: dto.StatusResponse{
			Status:  "success",
			Message: "PDF generated successfully",
		},
		OutputFilePath:  req.OutputFilePath,
		OutputFileBytes: generatePDFReq.OutputFileBytes,
	}, nil
}

func (s *PDFService) GeneratePDFStream(ctx context.Context, req *dto.PDFRequest) (*dto.GeneratePDFResponse, error) {
	startTime := time.Now()
	reqID := libutils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "GeneratePDFStream called :: ", map[string]any{"req_id": reqID})
	pdfSettings := &generateDoc.PDFParams{
		Landscape:           req.Landscape,
		DisplayHeaderFooter: false,
		PrintBackground:     true,
		PreferCssPageSize:   false,
		MarginTop:           req.MarginInch,
		MarginBottom:        req.MarginInch,
		MarginLeft:          req.MarginInch,
		MarginRight:         req.MarginInch,
		IsSinglePage:        req.SinglePage,
	}

	generatePDFReq := &generateDoc.PDFDto{
		ReqId:             reqID,
		InputTemplateUUID: req.TemplateUUID,
		Content:           req.Content,
		SignParams:        &generateDoc.SignParams{SignPdf: req.SignPdf},
		PdfParams:         pdfSettings,
	}

	if req.SignPdf {
		generatePDFReq.SignParams = &generateDoc.SignParams{
			SignPdf:       true,
			CertConfigKey: "digital_certificates.cert1",
		}
	}

	fileStorageAdapter, err := templatestore.TemplateStorageAdapterFactory(&templatestore.StorageConfig{StorageType: "stream"})
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in getting file storage adapter", err, nil)
		return nil, fmt.Errorf("failed to get file storage adapter: %w", err)
	}

	if err := generateDoc.GeneratePDF(ctx, generatePDFReq, s.TemplateStorageAdapter, &fileStorageAdapter); err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf stream", err, nil)
		return nil, fmt.Errorf("failed to generate PDF stream: %w", err)
	}

	if len(generatePDFReq.OutputFileBytes) == 0 {
		return nil, fmt.Errorf("no PDF data available")
	}

	duration := time.Since(startTime)
	svcUtils.Logger.Info(ctx, "generated pdf stream :: ", map[string]any{"req_id": reqID, "duration": duration})

	return &dto.GeneratePDFResponse{OutputFileBytes: generatePDFReq.OutputFileBytes}, nil
}

func (s *PDFService) SignPDF(ctx context.Context, req *dto.SignPDFRequest) (*dto.SignPDFResponse, error) {

	reqID := libutils.GenerateUniqueID(ctx)
	svcUtils.Logger.Info(ctx, "SignPDF called :: ", map[string]any{"req_id": reqID})

	if req.SignParams == nil || !req.SignParams.SignPdf {
		err := fmt.Errorf("sign_pdf must be true in sign_params")
		svcUtils.Logger.Error(ctx, "error in signing pdf", err, nil)
		return nil, err
	}

	signPDFDTO := &generateDoc.SignPDFDto{
		ReqId:          reqID,
		InputFilePath:  req.InputFilePath,
		InputFileBytes: req.InputFileBytes,
		OutputFilePath: req.OutputFilePath,
		SignParams:     req.SignParams,
	}

	if err := generateDoc.SignPDF(ctx, signPDFDTO, s.FileStorageAdapter); err != nil {
		svcUtils.Logger.Error(ctx, "error in signing pdf", err, nil)
		return nil, fmt.Errorf("failed to sign PDF: %w", err)
	}

	return &dto.SignPDFResponse{
		OutputFilePath:  req.OutputFilePath,
		OutputFileBytes: signPDFDTO.OutputFileBytes,
	}, nil
}
