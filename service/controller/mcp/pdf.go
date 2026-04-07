package tools

import (
	"context"

	"github.com/Zomato/espresso/service/dto"
	pdfservice "github.com/Zomato/espresso/service/service"
	svcUtils "github.com/Zomato/espresso/service/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PDFTools struct {
	service *pdfservice.PDFService
}

func NewPDFTools(service *pdfservice.PDFService) *PDFTools {
	return &PDFTools{service: service}
}

func (s *PDFTools) GeneratePDF(ctx context.Context, _ *mcp.CallToolRequest, req dto.GeneratePDFMCPRequest) (*mcp.CallToolResult, dto.GeneratePDFMCPResponse, error) {
	generatePDFReq, err := req.GeneratePDFMCPRequestToGeneratePDFRequest()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to parse generate PDF request: " + err.Error()}},
			IsError: true,
		}, dto.GeneratePDFMCPResponse{}, nil
	}

	svcResp, err := s.service.GeneratePDF(ctx, generatePDFReq)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in generating pdf", err, nil)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to generate PDF: " + err.Error()}},
			IsError: true,
		}, dto.GeneratePDFMCPResponse{}, nil
	}

	resp := svcResp.GeneratePDFResponseToGeneratePDFMCPResponse()
	return nil, *resp, nil
}

// func (s *PDFTools) GeneratePDFStream(ctx context.Context, _ *mcp.CallToolRequest, req dto.PDFMCPRequest) (*mcp.CallToolResult, any, error) {
// 	pdfReq, err := req.PDFMCPRequestToPDFRequest()
// 	if err != nil {
// 		return &mcp.CallToolResult{
// 			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to parse generate PDF stream request: " + err.Error()}},
// 			IsError: true,
// 		}, nil, nil
// 	}

// 	if err := pdfReq.Validate(); err != nil {
// 		return &mcp.CallToolResult{
// 			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
// 			IsError: true,
// 		}, nil, nil
// 	}

// 	resp, err := s.service.GeneratePDFStream(ctx, pdfReq)
// 	if err != nil {
// 		svcUtils.Logger.Error(ctx, "error in generating pdf stream", err, nil)
// 		return &mcp.CallToolResult{
// 			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to generate PDF stream: " + err.Error()}},
// 			IsError: true,
// 		}, nil, nil
// 	}

// 	fileName := "generated.pdf"
// 	if pdfReq.Filename != "" {
// 		fileName = pdfReq.Filename
// 		if !strings.HasSuffix(strings.ToLower(fileName), ".pdf") {
// 			fileName += ".pdf"
// 		}
// 	}
// 	fileName = filepath.Base(fileName)

// 	if len(resp.OutputFileBytes) == 0 {
// 		return &mcp.CallToolResult{
// 			Content: []mcp.Content{&mcp.TextContent{Text: "No PDF data available"}},
// 			IsError: true,
// 		}, nil, nil
// 	}

// 	return &mcp.CallToolResult{
// 		Content: []mcp.Content{
// 			&mcp.EmbeddedResource{
// 				Resource: &mcp.ResourceContents{
// 					URI:      "file://" + fileName,
// 					MIMEType: "application/pdf",
// 					Blob:     resp.OutputFileBytes,
// 				},
// 			},
// 		},
// 	}, nil, nil
// }

func (s *PDFTools) SignPDF(ctx context.Context, _ *mcp.CallToolRequest, req dto.SignPDFRequest) (*mcp.CallToolResult, dto.SignPDFResponse, error) {
	resp, err := s.service.SignPDF(ctx, &req)
	if err != nil {
		svcUtils.Logger.Error(ctx, "error in signing pdf", err, nil)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to sign PDF: " + err.Error()}},
			IsError: true,
		}, dto.SignPDFResponse{}, nil
	}

	return nil, *resp, nil
}
