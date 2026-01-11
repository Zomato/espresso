<div align="center">
  <img src="docs/assets/espresso.png" alt="Espresso Logo" width="260" height="115">
  <h1>Espresso: High Performance PDF Generator and Signer</h1>
</div>

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

- Go 1.22+
- Docker & Docker Compose (for running the complete service)
- X.509 certificates (for PDF signing)

---

## Logging

Espresso uses **Zero Logger** for structured, high-performance logging across all services. Logging is designed to provide clear observability into request handling, PDF generation, and signing operations while maintaining performance at scale.

### Logger Integration

The logger is initialized during application startup and injected into core components such as:

- HTTP request handlers
- PDF generation and signing workflows
- Storage backends
- Background workers

This ensures that logs are consistent, structured, and available across all layers of the system.

### Log Format

All logs are emitted in **JSON format** for easy parsing and ingestion into centralized logging systems.

Each log entry typically includes:

- `level` – Log severity (debug, info, warn, error)
- `time` – Timestamp of the event
- `message` – Human-readable description
- `component` – Logical component (e.g., api, pdf, signer, storage)
- `request_id` – Unique request identifier for traceability
- `event` – Business-level event name (e.g., pdf_generated, pdf_signed)

Example:

```json
{
  "level": "info",
  "time": "2026-01-12T10:15:30Z",
  "component": "pdf",
  "event": "pdf_generated",
  "request_id": "c3f21e8b",
  "message": "PDF generated successfully"
}
```

### Log Levels

Espresso uses standard log levels:

- **DEBUG** – Detailed internal execution information for troubleshooting.
- **INFO** – High-level business events such as request received, PDF generated, or document signed.
- **WARN** – Recoverable issues that do not stop execution (e.g., retries, degraded behavior).
- **ERROR** – Failures that impact request handling or system stability.

### Viewing Logs in Different Environments

**Development**  
When running locally with Docker Compose:

```bash
docker compose logs -f espresso
```

Logs are printed to stdout in JSON format.

**Staging**  
Logs are typically captured from container output and forwarded to the staging logging pipeline (e.g., file-based logs or centralized log collectors depending on infrastructure).

**Production**  
In production, logs are expected to be collected and indexed by a centralized logging system (e.g., ELK stack, Datadog, or similar).  
Structured fields such as `request_id`, `component`, and `event` enable efficient searching, alerting, and debugging at scale.

### Best Practices for Logging

When adding or modifying logs in Espresso:

- Do not log sensitive data such as certificates, private keys, PII, or raw document contents.
- Prefer structured fields over string concatenation.
- Log business events, not just technical steps (e.g., `pdf_generated`, `pdf_signed`).
- Include `request_id` wherever possible to allow tracing across services.
- Use appropriate log levels:
  - `INFO` for successful operations
  - `WARN` for recoverable issues
  - `ERROR` for failures
- Avoid excessive `DEBUG` logging in performance-critical paths.

### Adding New Logs

To add new logs:

1. Use the existing logger instance injected into the component.
2. Attach structured fields for context:
   - `component`
   - `request_id`
   - `event`
3. Keep messages concise and descriptive.

Example:

```go
logger.Info().
    Str("component", "signer").
    Str("event", "pdf_signed").
    Str("request_id", reqID).
    Msg("PDF signed successfully")
```

### Testing Log Output

Espresso includes tests that validate both functionality and observability.  
Where feasible, feature-level tests should assert that key workflows emit expected log events.  
For example:

- Generating a PDF should produce a log with `event=pdf_generated`
- Signing a document should emit `event=pdf_signed`

This ensures that logging remains consistent and reliable as the codebase evolves.

---

## Contributing

We welcome contributions!  
Please ensure that any new features include:

- Relevant documentation
- Appropriate logging
- Tests that validate both behavior and observability where applicable
