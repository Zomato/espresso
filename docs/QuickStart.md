# Quick Start Guide

This guide will help you run the Espresso service using Docker Compose.

## Prerequisites

- Docker and Docker Compose
- Make (optional)

## Running the Service

1. All required files are already provided in the repository:
   - HTML templates in `service/inputfiles/templates/`
   - Sample PDFs in `service/inputfiles/inputPDFs/`
   - Test certificates in `service/inputfiles/certificates/`
   - Configuration in `configs/espressoconfig.yaml`

2. Build and start all containers using make
```bash
make
```
or directly run docker compose from root
```bash
	DOCKERFILE=service/Dockerfile \
    docker-compose -f service/docker-compose.yml build && \
    docker-compose -f service/docker-compose.yml up -d 
```
then open http://localhost:3000 on a browser to access the espresso console

This will:
- Build all Docker containers
- Start MySQL for template storage
- Initialize LocalStack for S3 storage
- Launch the PDF service
- Start the web UI
- Open http://localhost:3000 in your browser

## Using the Web UI

1. **Template Management**:
   - Go to http://localhost:3000/templates
   - Click "Create New Template"
   - Fill in:
     - Template Name
     - HTML Content
     - JSON Schema (for form fields, placeholders, images, etc) 
   - Click "Save Template"

2. **Generate PDF**:
   - Go to http://localhost:3000/generate
   - Select a template
   - Input your JSON data
   - Click "Generate PDF"
   - Download or view the generated PDF

3. **Sign PDF**:
   - Go to http://localhost:3000/sign
   - Generate a PDF
   - Click "Generate signed PDF"
   - Download the signed PDF


## Troubleshooting

1. **Certificate Issues**:
   - Verify certificate files exist in the specified path
   - Check certificate password in config
   - Ensure certificate format is correct (X.509 for cert, PKCS#8 for key)

2. **Storage Issues**:
   - For disk storage: Check file permissions
   - For S3: Verify LocalStack is running (`docker ps`)
   - For MySQL: Check connection string in config

3. **Template Issues**:
   - Verify HTML is valid
   - Check JSON schema matches your data
   - Templates must be accessible from configured storage

4. **Common Errors**:
   - "Template not found": Check storage configuration
   - "Certificate load failed": Verify certificate paths and passwords
   - "PDF generation failed": Check HTML template validity
   - "Storage error": Verify storage configuration and permissions

## Logger Setup
The `ILogger` interface present inside `/lib/logger.go` supports the following log levels:

- `Info`
- `Warn`
- `Error`
- `Debug`

Each method accepts:

- `ctx`: A `context.Context` for request-scoped logging or tracing
- `msg`: Log message string
- `fields`: Key-value pairs for structured logging

#### Setting Up Logger

1. Implement the `ILogger` interface using your preferred logger.
2. Call `log.Initialize(yourLoggerInstance)` inside `/service/main.go`
3. Use `log.Logger.Info()`, `log.Logger.Error()`, etc., throughout your code.

#### Example: Using Zerolog
A sample implementation of ILogger interface can be found at `/service/utils/zerolog.go` and initialized in `service/main.go`

<br>

## Next Steps

- Check the API documentation for more endpoints
- Review the Integration Guide for programmatic usage
- Explore advanced features like custom headers/footers
- Configure production settings for deployment

For more detailed information, refer to the [Integration Guide](Integration.md).
