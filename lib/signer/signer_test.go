package signer

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignPdfStream(t *testing.T) {
	ctx := context.Background()

	// Generate test certificate and key
	cert, key := generateTestCertificate(t)

	tests := []struct {
		name    string
		pdfData []byte
		wantErr bool
	}{
		{
			name:    "valid_pdf",
			pdfData: getTestPDF(t),
			wantErr: false,
		},
		{
			name:    "invalid_pdf",
			pdfData: []byte("not a pdf"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader from the test PDF data
			pdfReader := bytes.NewReader(tt.pdfData)

			signedPDF, err := SignPdfStream(ctx, pdfReader, cert, key)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, signedPDF)
			assert.True(t, len(signedPDF) > len(tt.pdfData))

			// Verify it's still a valid PDF
			assert.True(t, bytes.HasPrefix(signedPDF, []byte("%PDF")))
		})
	}
}

// Helper functions
func generateTestCertificate(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Cert",
		},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(24 * time.Hour),
		SignatureAlgorithm: x509.SHA256WithRSA,
		KeyUsage:           x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, key
}

func getTestPDF(t *testing.T) []byte {
	// This is a minimal valid PDF file
	return []byte(`%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]>>endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000052 00000 n
0000000101 00000 n
trailer<</Size 4/Root 1 0 R>>
startxref
163
%%EOF`)
}
