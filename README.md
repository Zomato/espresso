# Espresso: High Performance PDF Generator and Signer

Espresso is the ultimate solution for high-performance PDF generation and digital signing. Whether you need to generate PDFs from HTML templates or sign them with digital certificates, Espresso is designed to handle massive workloads with ease. With rendering and signing times under 200ms, Espresso is ready to handle peak loads of 120K requests per minute (RPM).

We recently signed 1.6 million PDFs in just 19 minutes—that’s ~1,400 PDFs per second. 


## Key Features

- **High Performance**: 
  - PDF Generation: < 200ms per document
  - Digital Signing: > 1,400 PDFs/second
  - Production tested at 120K RPM

- **Core Capabilities**:
  - HTML to PDF conversion with full CSS support
  - Digital signing with X.509 certificates
  - Multiple storage backends for templates (S3, MySQL, Disk)
  - REST API interface
  - Browser-based template management UI


## Quick Start

See our [Quick Start Guide](docs/QuickStart.md) for running the service using Docker Compose.


## Requirements

- Go 1.23+
- Docker & Docker Compose (for running the complete service)
- X.509 certificates (for PDF signing)

