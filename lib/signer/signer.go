package signer

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/digitorus/pdf"
	"github.com/digitorus/pkcs7"
	"github.com/mattetti/filebuffer"
)

func Sign(input io.ReadSeeker, output io.Writer, rdr *pdf.Reader, size int64, sign_data SignData) error {
	sign_data.objectId = uint32(rdr.XrefInformation.ItemCount) + 2

	context := SignContext{
		PDFReader:              rdr,
		InputFile:              input,
		OutputFile:             output,
		SignData:               sign_data,
		SignatureMaxLengthBase: uint32(hex.EncodedLen(512)),
	}

	existingSignatures, err := context.fetchExistingSignatures()
	if err != nil {
		return err
	}
	context.existingSignatures = existingSignatures

	err = context.SignPDF()
	if err != nil {
		return err
	}

	return nil
}

func (context *SignContext) SignPDF() error {

	if context.SignData.Signature.CertType == 0 {
		context.SignData.Signature.CertType = 1
	}
	if context.SignData.Signature.DocMDPPerm == 0 {
		context.SignData.Signature.DocMDPPerm = 1
	}
	if !context.SignData.DigestAlgorithm.Available() {
		context.SignData.DigestAlgorithm = crypto.SHA256
	}
	if context.SignData.Appearance.Page == 0 {
		context.SignData.Appearance.Page = 1
	}

	context.OutputBuffer = filebuffer.New([]byte{})

	_, err := context.InputFile.Seek(0, 0)
	if err != nil {
		return err
	}
	if _, err := io.Copy(context.OutputBuffer, context.InputFile); err != nil {
		return err
	}

	if _, err := context.OutputBuffer.Write([]byte("\n")); err != nil {
		return err
	}

	context.SignatureMaxLength = context.SignatureMaxLengthBase

	if context.SignData.Signature.CertType != TimeStampSignature {
		switch context.SignData.Certificate.SignatureAlgorithm.String() {
		case "SHA1-RSA":
		case "ECDSA-SHA1":
		case "DSA-SHA1":
			context.SignatureMaxLength += uint32(hex.EncodedLen(128))
		case "SHA256-RSA":
		case "ECDSA-SHA256":
		case "DSA-SHA256":
			context.SignatureMaxLength += uint32(hex.EncodedLen(256))
		case "SHA384-RSA":
		case "ECDSA-SHA384":
			context.SignatureMaxLength += uint32(hex.EncodedLen(384))
		case "SHA512-RSA":
		case "ECDSA-SHA512":
			context.SignatureMaxLength += uint32(hex.EncodedLen(512))
		}

		context.SignatureMaxLength += uint32(hex.EncodedLen(context.SignData.DigestAlgorithm.Size() * 2))

		degenerated, err := pkcs7.DegenerateCertificate(context.SignData.Certificate.Raw)
		if err != nil {
			return fmt.Errorf("failed to degenerate certificate: %w", err)
		}

		context.SignatureMaxLength += uint32(hex.EncodedLen(len(degenerated)))

		context.SignatureMaxLength += uint32(hex.EncodedLen(len(context.SignData.Certificate.RawIssuer)))

		var certificate_chain []*x509.Certificate
		if len(context.SignData.CertificateChains) > 0 && len(context.SignData.CertificateChains[0]) > 1 {
			certificate_chain = context.SignData.CertificateChains[0][1:]
		}

		if len(certificate_chain) > 0 {
			for _, cert := range certificate_chain {
				degenerated, err := pkcs7.DegenerateCertificate(cert.Raw)
				if err != nil {
					return fmt.Errorf("failed to degenerate certificate in chain: %w", err)
				}

				context.SignatureMaxLength += uint32(hex.EncodedLen(len(degenerated)))
			}
		}

		if err := context.fetchRevocationData(); err != nil {
			return fmt.Errorf("failed to fetch revocation data: %w", err)
		}
	}

	if context.SignData.TSA.URL != "" {
		context.SignatureMaxLength += uint32(hex.EncodedLen(9000))
	}

	var signature_object []byte

	switch context.SignData.Signature.CertType {
	case TimeStampSignature:
		signature_object = context.createTimestampPlaceholder()
	default:
		signature_object = context.createSignaturePlaceholder()
	}

	context.SignData.objectId, err = context.addObject(signature_object)
	if err != nil {
		return fmt.Errorf("failed to add signature object: %w", err)
	}

	visible := false
	rectangle := [4]float64{0, 0, 0, 0}
	if context.SignData.Signature.CertType != ApprovalSignature && context.SignData.Appearance.Visible {
		return fmt.Errorf("visible signatures are only allowed for approval signatures")
	} else if context.SignData.Signature.CertType == ApprovalSignature && context.SignData.Appearance.Visible {
		visible = true
		rectangle = [4]float64{
			context.SignData.Appearance.LowerLeftX,
			context.SignData.Appearance.LowerLeftY,
			context.SignData.Appearance.UpperRightX,
			context.SignData.Appearance.UpperRightY,
		}
	}

	visual_signature, err := context.createVisualSignature(visible, context.SignData.Appearance.Page, rectangle)
	if err != nil {
		return fmt.Errorf("failed to create visual signature: %w", err)
	}

	context.VisualSignData.objectId, err = context.addObject(visual_signature)
	if err != nil {
		return fmt.Errorf("failed to add visual signature object: %w", err)
	}

	if context.SignData.Appearance.Visible {
		inc_page_update, err := context.createIncPageUpdate(context.SignData.Appearance.Page, context.VisualSignData.objectId)
		if err != nil {
			return fmt.Errorf("failed to create incremental page update: %w", err)
		}
		err = context.updateObject(context.VisualSignData.pageObjectId, inc_page_update)
		if err != nil {
			return fmt.Errorf("failed to add incremental page update object: %w", err)
		}
	}

	catalog, err := context.createCatalog()
	if err != nil {
		return fmt.Errorf("failed to create catalog: %w", err)
	}

	context.CatalogData.ObjectId, err = context.addObject(catalog)
	if err != nil {
		return fmt.Errorf("failed to add catalog object: %w", err)
	}

	if err := context.writeXref(); err != nil {
		return fmt.Errorf("failed to write xref: %w", err)
	}

	if err := context.writeTrailer(); err != nil {
		return fmt.Errorf("failed to write trailer: %w", err)
	}

	if err := context.updateByteRange(); err != nil {
		return fmt.Errorf("failed to update byte range: %w", err)
	}

	if err := context.replaceSignature(); err != nil {
		return fmt.Errorf("failed to replace signature: %w", err)
	}

	if _, err := context.OutputBuffer.Seek(0, 0); err != nil {
		return err
	}
	file_content := context.OutputBuffer.Buff.Bytes()

	if _, err := context.OutputFile.Write(file_content); err != nil {
		return err
	}

	return nil
}

func (context *SignContext) fetchExistingSignatures() ([]SignData, error) {
	var signatures []SignData

	acroForm := context.PDFReader.Trailer().Key("Root").Key("AcroForm")
	if acroForm.IsNull() {
		return signatures, nil
	}

	fields := acroForm.Key("Fields")
	if fields.IsNull() {
		return signatures, nil
	}

	for i := 0; i < fields.Len(); i++ {
		field := fields.Index(i)
		if field.Key("FT").Name() == "Sig" {
			ptr := field.GetPtr()
			sig := SignData{
				objectId: uint32(ptr.GetID()),
			}
			signatures = append(signatures, sig)
		}
	}

	return signatures, nil
}

func (context *SignContext) createPropBuild() string {
	var buffer bytes.Buffer

	buffer.WriteString(" /Prop_Build <<\n")
	buffer.WriteString("   /App << /Name /Zomato-PDF-Signer >>\n")
	buffer.WriteString(" >>\n")

	return buffer.String()
}

func SignPdfStream(ctx context.Context, pdfStream io.Reader, cert *x509.Certificate, privateKey crypto.Signer) ([]byte, error) {

	var pdfBuffer bytes.Buffer
	_, err := io.Copy(&pdfBuffer, pdfStream)
	if err != nil {
		return nil, fmt.Errorf("failed to read pdf stream: %v", err)
	}

	pdfBytes := pdfBuffer.Bytes()

	pdfReader, err := pdf.NewReader(bytes.NewReader(pdfBytes), int64(len(pdfBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %v", err)
	}

	outputBuffer := new(bytes.Buffer)

	inputPdf := bytes.NewReader(pdfBytes)
	size := int64(len(pdfBytes))

	err = Sign(inputPdf, outputBuffer, pdfReader, size, SignData{
		Signature: SignDataSignature{
			Info: SignDataSignatureInfo{
				Name:        "John Doe",
				Location:    "locdummt",
				Reason:      "reason",
				ContactInfo: "https://www.org.com",
				Date:        time.Now().Local(),
			},
			CertType:   CertificationSignature,
			DocMDPPerm: AllowFillingExistingFormFieldsAndSignaturesPerms,
		},
		Signer:            privateKey,
		DigestAlgorithm:   crypto.SHA256,
		Certificate:       cert,
		CertificateChains: [][]*x509.Certificate{{cert}},
		TSA:               TSA{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign PDF: %v", err)
	}

	return outputBuffer.Bytes(), nil
}
