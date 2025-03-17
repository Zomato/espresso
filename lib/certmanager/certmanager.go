package certmanager

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
)

// SigningCredentials holds the certificate and private key for PDF signing
type SigningCredentials struct {
	Certificate *x509.Certificate
	PrivateKey  crypto.Signer
}
type CertificateConfig struct {
	CertFilePath string
	KeyFilePath  string
	KeyPassword  string
}

// LoadSigningCredentials loads certificate and private key from configured paths
func LoadSigningCredentials(ctx context.Context, certConfig *CertificateConfig) (*SigningCredentials, error) {

	cert, err := getCertificate(certConfig.CertFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %v", err)
	}

	privateKey, err := getKey(certConfig.KeyFilePath, certConfig.KeyPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %v", err)
	}

	return &SigningCredentials{
		Certificate: cert,
		PrivateKey:  privateKey,
	}, nil
}

func getCertificate(certPath string) (*x509.Certificate, error) {

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %v", err)
	}

	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}
	return cert, nil
}

func getKey(keyPath, password string) (crypto.Signer, error) {

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	privateKey, err := loadPrivateKey(keyBytes, password)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	return privateKey, nil
}

func loadPrivateKey(keyBytes []byte, password string) (crypto.Signer, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type == "ENCRYPTED PRIVATE KEY" {
		// For PKCS#8 encrypted keys
		key, err := parseEncryptedPKCS8PrivateKey(block.Bytes, []byte(password))
		if err != nil {
			return nil, fmt.Errorf("failed to parse encrypted PKCS#8 private key: %w", err)
		}

		switch k := key.(type) {
		case *rsa.PrivateKey:
			return k, nil
		case *ecdsa.PrivateKey:
			return k, nil
		default:
			return nil, fmt.Errorf("unsupported private key type")
		}
	} else if block.Type == "PRIVATE KEY" {
		// For unencrypted PKCS#8 keys
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
		}

		switch k := key.(type) {
		case *rsa.PrivateKey:
			return k, nil
		case *ecdsa.PrivateKey:
			return k, nil
		default:
			return nil, fmt.Errorf("unsupported private key type")
		}
	} else {
		return nil, fmt.Errorf("unsupported private key format: %s", block.Type)
	}
}

// Helper function to parse PKCS#8 encrypted private keys
func parseEncryptedPKCS8PrivateKey(data, password []byte) (interface{}, error) {
	var privKey interface{}

	// First try direct parsing (for some implementations)
	privKey, err := x509.ParsePKCS8PrivateKey(data)
	if err == nil {
		return privKey, nil
	}

	tmpFile, err := os.CreateTemp("", "encrypted_key_*.pem")
	if err != nil {
		return nil, err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	err = os.WriteFile(tmpName, pem.EncodeToMemory(&pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: data,
	}), 0600)
	if err != nil {
		return nil, err
	}

	// Create another temp file for the decrypted key
	decryptedTmpFile, err := os.CreateTemp("", "decrypted_key_*.pem")
	if err != nil {
		return nil, err
	}
	decryptedName := decryptedTmpFile.Name()
	defer os.Remove(decryptedName)

	// Use OpenSSL to decrypt the key
	cmd := exec.Command("openssl", "pkcs8", "-in", tmpName, "-out", decryptedName, "-passin", "pass:"+string(password))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("OpenSSL error: %v, %s", err, stderr.String())
	}

	// Read the decrypted key
	decryptedData, err := os.ReadFile(decryptedName)
	if err != nil {
		return nil, err
	}

	decryptedBlock, _ := pem.Decode(decryptedData)
	if decryptedBlock == nil {
		return nil, fmt.Errorf("failed to decode decrypted PEM block")
	}

	return x509.ParsePKCS8PrivateKey(decryptedBlock.Bytes)
}
