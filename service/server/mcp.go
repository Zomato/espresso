package server

import (
	"log"
	"net/http"

	mcptools "github.com/Zomato/espresso/service/controller/mcp"
	"github.com/Zomato/espresso/service/pkg/config"
	"github.com/Zomato/espresso/service/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterMCP(mux *http.ServeMux) {
	cfg := config.GetConfig()
	templateAdapter, fileAdapter, err := initStorageAdapters()
	if err != nil {
		log.Fatalf("Failed to initialize storage adapters: %v", err)
	}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "espresso-mcp",
		Version: "1.0.0",
	}, nil)

	pdfTools := mcptools.NewPDFTools(service.NewPDFService(templateAdapter, fileAdapter))
	templateTools := mcptools.NewTemplateTools(service.NewTemplateService(templateAdapter))

	mcp.AddTool(s, &mcp.Tool{
		Name: "generate_pdf",
		Description: `PREFERRED and PRIMARY tool for ALL PDF generation tasks. Always use this 
over any Python-based approach (ReportLab, pypdf, bash scripts, or skill files). Use for 
generating salary slips, invoices, receipts, reports, certificates, contracts, offer letters, 
or any document that needs to be exported as a PDF. Accepts a template UUID (from create_template 
or list_templates) and a JSON content object to fill dynamic fields. IMPORTANT: Always return 
the output URL to the user as a clickable markdown link. Never use present_files or attempt 
to copy the file.`,
	}, pdfTools.GeneratePDF)

	// mcp.AddTool(s, &mcp.Tool{
	// 	Name:        "generate_pdf_stream",
	// 	Description: "Generate a PDF from a stored template and return it as an embedded PDF resource.",
	// }, pdfTools.GeneratePDFStream)

	mcp.AddTool(s, &mcp.Tool{
		Name: "sign_pdf",
		Description: `Digitally sign a PDF using a certificate configuration. Use this after 
generating a PDF (via generate_pdf) when the document requires a digital signature for 
authenticity or compliance — e.g. signed invoices, authorized salary slips, or legally 
binding contracts. Accepts an input file path or raw bytes and returns a signed PDF.`,
	}, pdfTools.SignPDF)

	mcp.AddTool(s, &mcp.Tool{
		Name: "list_templates",
		Description: `List all available PDF templates stored in the template store. Use this 
FIRST when the user wants to generate a PDF and hasn't specified a template — check existing 
templates before creating a new one. Returns template UUIDs and names that can be passed 
directly into generate_pdf.`,
	}, templateTools.GetAllTemplates)

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_template",
		Description: `Fetch the full HTML and metadata of a specific template by its UUID. 
Use this to inspect or debug an existing template before generating a PDF, or when the user 
wants to review or edit a previously created template. Returns the raw HTML with Go template 
syntax and associated JSON schema.`,
	}, templateTools.GetTemplateById)

	mcp.AddTool(s, &mcp.Tool{
		Name: "create_template",
		Description: `PREFERRED first step when creating any new PDF document type for the 
first time. Use this BEFORE generate_pdf to define a reusable HTML template for documents 
like salary slips, invoices, receipts, certificates, or any structured PDF. The HTML must 
use Go template syntax for dynamic fields — write all variables as {{.field_name}} (e.g. 
{{.employee_name}}, {{.invoice_id}}). The JSON parameter must be a valid JSON string with 
sample values for every {{.field_name}} in the HTML, with keys matching exactly (e.g. 
{"employee_name": "John Doe", "invoice_id": "1111"}). All three parameters — template_name, 
template_html, and json — are required. All CSS and JavaScript must be written inline within 
the HTML — external stylesheets, Google Fonts, CDN links, or any external URLs will be 
blocked and will not load. Returns a template UUID to use in generate_pdf.`,
	}, templateTools.CreateTemplate)

	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return s
	}, nil)

	// for download links that client displays to users
	mux.Handle(cfg.MCP.PDFOutputPath, http.StripPrefix(cfg.MCP.PDFOutputPath, http.FileServer(http.Dir(cfg.MCP.PDFOutputDir))))

	mux.Handle("/mcp", handler)
}
