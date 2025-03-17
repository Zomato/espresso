package signer

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/crypto/ocsp"
)

func (context *SignContext) fetchRevocationData() error {
	if context.SignData.RevocationFunction != nil {
		if context.SignData.CertificateChains != nil && (len(context.SignData.CertificateChains) > 0) {
			certificate_chain := context.SignData.CertificateChains[0]
			if certificate_chain != nil && (len(certificate_chain) > 0) {
				for i, certificate := range certificate_chain {
					if i < len(certificate_chain)-1 {
						err := context.SignData.RevocationFunction(certificate, certificate_chain[i+1], &context.SignData.RevocationData)
						if err != nil {
							return err
						}
					} else {
						err := context.SignData.RevocationFunction(certificate, nil, &context.SignData.RevocationData)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	for _, crl := range context.SignData.RevocationData.CRL {
		context.SignatureMaxLength += uint32(hex.EncodedLen(len(crl.FullBytes)))
	}
	for _, ocsp := range context.SignData.RevocationData.OCSP {
		context.SignatureMaxLength += uint32(hex.EncodedLen(len(ocsp.FullBytes)))
	}

	return nil
}

func DefaultEmbedRevocationStatusFunction(cert, issuer *x509.Certificate, i *InfoArchival) error {

	if issuer != nil && len(cert.OCSPServer) > 0 {
		err := embedOCSPRevocationStatus(cert, issuer, i)
		if err != nil {
			return err
		}
	}

	if len(cert.CRLDistributionPoints) > 0 {
		err := embedCRLRevocationStatus(cert, issuer, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func embedOCSPRevocationStatus(cert, issuer *x509.Certificate, i *InfoArchival) error {
	req, err := ocsp.CreateRequest(cert, issuer, nil)
	if err != nil {
		return err
	}

	ocspUrl := fmt.Sprintf("%s/%s", strings.TrimRight(cert.OCSPServer[0], "/"),
		base64.StdEncoding.EncodeToString(req))

	resp, err := http.Get(ocspUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	_, err = ocsp.ParseResponseForCert(body, cert, issuer)
	if err != nil {
		return err
	}

	return i.AddOCSP(body)
}

func embedCRLRevocationStatus(cert, issuer *x509.Certificate, i *InfoArchival) error {
	resp, err := http.Get(cert.CRLDistributionPoints[0])
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return i.AddCRL(body)
}

func (r *InfoArchival) AddCRL(b []byte) error {
	r.CRL = append(r.CRL, asn1.RawValue{FullBytes: b})
	return nil
}

func (r *InfoArchival) AddOCSP(b []byte) error {
	r.OCSP = append(r.OCSP, asn1.RawValue{FullBytes: b})
	return nil
}
