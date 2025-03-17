# Espresso: Go PDF Generation and Signing Library

Espresso library is a high-performance library for generating PDFs from HTML templates and digitally signing them. With rendering times under 200ms, it's built for high-throughput applications.

## Installation

```bash
go get github.com/Zomato/espresso/lib
```

## Dependencies Setup

Here's an example dockerfile to set up all required dependencies:

```dockerfile
FROM --platform=$BUILDPLATFORM golang:1.22.4-bullseye

WORKDIR /app/example/



# Configure
ENV GO111MODULE=on \
    ROD_BROWSER_BIN=/usr/bin/chromium

# Install browser dependencies, Chromium, and netcat
RUN apt-get update && apt-get install -y \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libnss3 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libdrm2 \
    libgbm1 \
    libxshmfence1 \
    libx11-xcb1 \
    chromium \
    netcat \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Add a non-root user for running Chrome with sandbox
RUN groupadd -r chrome && useradd -r -g chrome -G audio,video chrome \
    && mkdir -p /home/chrome/Downloads \
    && chown -R chrome:chrome /home/chrome \
    && chown -R chrome:chrome /app/example


# Set proper permissions
RUN chown -R chrome:chrome /app

# Set the user to chrome for the container
USER chrome

EXPOSE 8081

# Change the CMD as per your code
CMD ["go", "run","-mod=mod", "/app/example/main.go"]

```

## Basic Usage

### 1. Initialize Browser and Worker Pool for PDF generation

First, initialize the browser manager and worker pool. This is required before generating any PDFs:

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/Zomato/espresso/lib/browser_manager"
    "github.com/Zomato/espresso/lib/workerpool"
)

func main() {
    ctx := context.Background()
    
    // Initialize browser manager
    tabPoolSize := 5 // number of concurrent browser tabs
    if err := browser_manager.Init(ctx, tabPoolSize); err != nil {
        log.Fatalf("Failed to initialize browser: %v", err)
    }

    // Initialize worker pool
    workerCount := 10 // number of concurrent workers
    workerTimeout := 200 // timeout in milliseconds
    workerpool.Initialize(
        workerCount,
        time.Duration(workerTimeout) * time.Millisecond,
    )
}
```

### 2. PDF Generation

Here's a basic example of generating a PDF from HTML:

```go
package main

import (
    "context"
    "github.com/Zomato/espresso/lib/renderer"
    "github.com/Zomato/espresso/lib/browser_manager"
    "github.com/go-rod/rod/lib/proto"
)

func main() {
    ctx := context.Background()
    // initialize browser and workerpool here

    // Configure viewport (optional)
    viewport := &browser_manager.ViewportConfig{
        Width: 794,             // A4 width
        Height: 1124,           // A4 height
        DeviceScaleFactor: 1.0,
        IsMobile: false,
    }
    // Configure PDF settings
    pdfSettings := &proto.PagePrintToPDF{
        Landscape: false,
        PrintBackground: true,
        PreferCSSPageSize: false,
        MarginTop: float64Ptr(0.4),    // inches
        MarginBottom: float64Ptr(0.4),  // inches
        MarginLeft: float64Ptr(0.4),    // inches
        MarginRight: float64Ptr(0.4),   // inches
        // Optional settings:
        // PaperWidth: float64Ptr(8.27),  // A4 width in inches
        // PaperHeight: float64Ptr(11.7), // A4 height in inches
        // DisplayHeaderFooter: true,
        // HeaderTemplate: "<div>Custom Header</div>",
        // FooterTemplate: "<div>Page <span class='pageNumber'></span></div>",
        // PageRanges: "1-2",             // Specific pages to print
    }

    // Generate PDF
    input := &renderer.GetHtmlPdfInput{
        Data: []byte(`{"title": "Hello", "content": "World"}`),
        ViewPort: viewport,
        PdfParams: pdfSettings,
        TemplateRequest: templatestore.GetTemplateRequest{
            // TemplateUUID: "your-template-id", // If using mysql template storage, make sure to init and pass the mysql storage adapter to GetHtmlPdf
            // Or directly provide HTML content:
            // TemplatePath: "/path/to/your-template", // for disk storage
	        // TemplateS3Path: "/your/s3bucket/key", // for s3 storage
            TemplateBytes: []byte(`
                <html>
                    <body>
                        <h1>{{.title}}</h1>
                        <p>{{.content}}</p>
                    </body>
                </html>
            `),
        },
    }

    pdf, err := renderer.GetHtmlPdf(ctx, input, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer pdf.Close()

    // Use the PDF stream
    io.Copy(outputFile, pdf)
}

func float64Ptr(v float64) *float64 {
    return &v
}
```
Note- If using any of the s3, mysql, disk storage adapters, make sure to init and pass the adapters as a parameter in GetHtmlPdf [See configuration example](#L147)

### 3. PDF Signing (Basic)

```go
package main

import (
    "context"
    "io"
    "github.com/Zomato/espresso/lib/signer"
)

func main() {
    ctx := context.Background()
    
    // Your PDF stream
    pdfStream := getPDFStream() // io.Reader
    
    // Your certificate and private key
    // var cert *x509.Certificate
    // var privateKey crypto.Signer
    
    // Sign PDF
    signedPDF, err := signer.SignPdfStream(ctx, pdfStream, cert, privateKey)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the signed PDF bytes
    io.Copy(outputFile, bytes.NewReader(signedPDF))
}
```

## Important Parameters

### Viewport Configuration
- `Width`: Width in pixels (default: 794 for A4)
- `Height`: Height in pixels (default: 1124 for A4)
- `DeviceScaleFactor`: Scale factor for rendering (default: 1.0)
- `IsMobile`: Whether to use mobile rendering (default: false)

### PDF Settings
- `Landscape`: Toggle landscape orientation
- `PrintBackground`: Include background graphics (default: true)
- `MarginTop/Bottom/Left/Right`: Margins in inches
- `PaperWidth/Height`: Paper dimensions in inches
- `DisplayHeaderFooter`: Enable custom headers/footers
- `HeaderTemplate/FooterTemplate`: HTML templates for headers/footers
- `PageRanges`: Specify pages to include (e.g., "1-5")
- `PreferCSSPageSize`: Use CSS page size over paper size

### Template Variables
- Templates use Go's text/template syntax
- Data is passed as JSON and mapped to template variables
- Access variables using `{{.variableName}}`

## Storage Adapters

lib supports multiple storage adapters for templates and generated PDFs:

### 1. Direct HTML (No Storage Adapter)
```go
input := &renderer.GetHtmlPdfInput{
    TemplateRequest: templatestore.GetTemplateRequest{
        TemplateBytes: []byte(`<html><body>Hello {{.name}}</body></html>`),
    },
    Data: []byte(`{"name": "World"}`),
}

pdf, err := renderer.GetHtmlPdf(ctx, input, nil) // Note: nil adapter
```

### 2. Disk Storage
```go

diskAdapter, err := templatestore.TemplateStorageAdapterFactory(templatestore.TemplateStorageConfig{
    StorageType: "disk",
})
input := &renderer.GetHtmlPdfInput{
    TemplateRequest: templatestore.GetTemplateRequest{
        TemplatePath: "/path/to/your/template.html", // Must be accessible
    },
    Data: []byte(`{"name": "World"}`),
}

pdf, err := renderer.GetHtmlPdf(ctx, input, &diskAdapter)
```

### 3. S3 Storage
Note: You can use your own implementation to replace the spf13/viper package
```go
// Required config:
// s3:
//   endpoint: "https://s3.amazonaws.com"
//   region: "us-west-2"
//   bucket: "your-bucket"
// aws:
//   accessKeyID: "your-key"
//   secretAccessKey: "your-secret"

s3Adapter, err := templatestore.TemplateStorageAdapterFactory(templatestore.TemplateStorageConfig{
    StorageType: "s3",
    S3Config   :  &s3.Config{
        // your configuration
        Endpoint:              viper.GetString("s3.endpoint"),
		Region:                viper.GetString("s3.region"),
		Bucket:                viper.GetString("s3.bucket"),
		Debug:                 viper.GetBool("s3.debug"),
		ForcePathStyle:        viper.GetBool("s3.forcePathStyle"),
		UploaderConcurrency:   viper.GetInt("s3.uploaderConcurrency"),
		UploaderPartSize:      viper.GetInt64("s3.uploaderPartSize"),
		DownloaderConcurrency: viper.GetInt("s3.downloaderConcurrency"),
		DownloaderPartSize:    viper.GetInt64("s3.downloaderPartSize"),
		RetryMaxAttempts:      viper.GetInt("s3.retryMaxAttempts"),
		UseCustomTransport:    viper.GetBool("s3.useCustomTransport"),
    }
	AwsCredConfig &s3.AwsCredConfig{
        // your aws creds
        AccessKeyID:     viper.GetString("aws.accessKeyID"),
		SecretAccessKey: viper.GetString("aws.secretAccessKey"),
		SessionToken:    viper.GetString("aws.sessionToken"),
    }
})
input := &renderer.GetHtmlPdfInput{
    TemplateRequest: templatestore.GetTemplateRequest{
        TemplateS3Path: "templates/mytemplate.html",
    },
    Data: []byte(`{"name": "World"}`),
}

pdf, err := renderer.GetHtmlPdf(ctx, input, &s3Adapter)
```

### 4. MySQL Storage
```go

mysqlAdapter, err := templatestore.TemplateStorageAdapterFactory(templatestore.TemplateStorageConfig{
    StorageType: "mysql",
    MysqlDSN: "your mysql dsn connection string"
})
input := &renderer.GetHtmlPdfInput{
    TemplateRequest: templatestore.GetTemplateRequest{
        TemplateUUID: "template-1-uuid", // UUID from your templates table
    },
    Data: []byte(`{"name": "World"}`),
}

pdf, err := renderer.GetHtmlPdf(ctx, input, &mysqlAdapter)
```

## Digital Signing in Detail

lib includes a robust certificate manager for PDF signing. Here's a detailed guide:

### Certificate Requirements

1. X.509 Certificate format
2. Supported private key formats:
   - PKCS#8 encrypted private key
   - PKCS#8 unencrypted private key
   - RSA and ECDSA keys supported

### Using CertManager

```go
// 1. Using certificates
certConfig := &certmanager.CertificateConfig{
			CertFilePath: "/path/to/cert",
			KeyFilePath:  "/path/to/private/key",
			KeyPassword:  "optional-for password protected private key",
		}
credentials, err := certmanager.LoadSigningCredentials(ctx, certConfig)
signedPDF, err := signer.SignPdfStream(ctx, pdfStream, credentials.Certificate, credentials.PrivateKey)

// 2. Direct certificate usage
import (
    "crypto/x509"
    "crypto"
)

// Your certificate and private key loading logic
cert *x509.Certificate
privateKey crypto.Signer

signedPDF, err := signer.SignPdfStream(ctx, pdfStream, cert, privateKey)
```

### Example Certificate Format
```
# Certificate (cert.pem)
-----BEGIN CERTIFICATE-----
MIICxxxxxxxxxxxxxxxxxxxxxx
-----END CERTIFICATE-----

# Encrypted Private Key (key_pkcs8_encrypted.pem)
-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIFxxxxxxxxxxxxxxxxxxxxxx
-----END ENCRYPTED PRIVATE KEY-----
```

### Generating your own self-signed certificate using openssl

```bash
openssl genpkey -algorithm RSA -out key_unencrypted.pem -pkeyopt rsa_keygen_bits:2048
openssl pkcs8 -in key_unencrypted.pem -topk8 -out key_pkcs8_encrypted.pem -v2 aes-256-cbc
# Password: test123
openssl req -new -key key_pkcs8_encrypted.pem -out cert.csr -subj "/CN=Test"
# Password: test123
openssl x509 -req -in cert.csr -signkey key_pkcs8_encrypted.pem -out cert.pem -days 365 -extfile <(printf "extendedKeyUsage=codeSigning")
# Password: test123
```

## Running Tests

The project has a comprehensive test suite covering both unit tests for the lib library and integration tests for the service.

### Test Structure

1. **lib Unit Tests**
   - `lib/renderer/renderer_test.go`: Tests HTML to PDF conversion with various templates and configurations
   - `lib/signer/signer_test.go`: Tests PDF digital signing with test certificates

2. **Service Integration Tests**
   - `service/integration_test.go`: End-to-end API tests covering PDF generation and template management endpoints

### Running Tests Using Make

The project includes several make targets for running tests:

```bash
# Run all tests (both lib and service)
make test

# Run only lib library tests
make test-lib

# Run only service integration tests
make test-service

# Clean up test environment
make clean
```

### Writing New Tests

When adding new tests, follow these patterns:

1. Use table-driven tests for multiple scenarios:
```go
tests := []struct {
    name    string
    input   interface{}
    wantErr bool
    isPDF   bool      // for response type checking
}{
    // test cases
}
```

2. Include comprehensive test cases:
   - Happy path scenarios
   - Error cases
   - Edge cases
   - Different input configurations

3. Verify PDF output:
```go
assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
```

4. Clean up resources:
```go
defer resp.Body.Close()
// or
defer pdf.Close()
```

## Best Practices

1. **Resource Management**
   - Always close PDF streams using `defer pdf.Close()`
   - Use appropriate viewport sizes for your content

2. **Performance**
   - Reuse template instances when possible
   - Consider caching frequently used templates
   - Use appropriate image formats and sizes

3. **Security**
   - Validate all template input
   - Use secure certificate storage for signing
   - Implement proper access controls

## For More Examples

Check out our example service implementation in the `service` directory, which showcases a complete web server using this package.

## License

This project is licensed under the MIT License - see the LICENSE file for details.