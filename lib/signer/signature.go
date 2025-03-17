package signer

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digitorus/pkcs7"
	"github.com/digitorus/timestamp"
	"golang.org/x/crypto/cryptobyte"

	cryptobyte_asn1 "golang.org/x/crypto/cryptobyte/asn1"
)

func (context *SignContext) createSignaturePlaceholder() []byte {

	var signature_buffer bytes.Buffer

	signature_buffer.WriteString("<<\n")
	signature_buffer.WriteString(" /Type /Sig\n")
	signature_buffer.WriteString(" /Filter /Adobe.PPKLite\n")
	signature_buffer.WriteString(" /SubFilter /adbe.pkcs7.detached\n")

	signature_buffer.WriteString(context.createPropBuild())

	signature_buffer.WriteString(" " + signatureByteRangePlaceholder)

	signature_buffer.WriteString(" /Contents<")
	signature_buffer.Write(bytes.Repeat([]byte("0"), int(context.SignatureMaxLength)))
	signature_buffer.WriteString(">\n")

	switch context.SignData.Signature.CertType {
	case CertificationSignature, UsageRightsSignature:
		signature_buffer.WriteString(" /Reference [\n")
		signature_buffer.WriteString(" << /Type /SigRef\n")
	}

	switch context.SignData.Signature.CertType {

	case CertificationSignature:
		signature_buffer.WriteString(" /TransformMethod /DocMDP\n")

		signature_buffer.WriteString(" /TransformParams <<\n")

		signature_buffer.WriteString("   /Type /TransformParams\n")

		signature_buffer.WriteString("   /P " + strconv.Itoa(int(context.SignData.Signature.DocMDPPerm)))

		signature_buffer.WriteString("   /V /1.2\n")

	case UsageRightsSignature:
		signature_buffer.WriteString("   /TransformMethod /UR3\n")

		signature_buffer.WriteString("   /TransformParams <<\n")
		signature_buffer.WriteString("     /Type /TransformParams\n")
		signature_buffer.WriteString("     /V /2.2\n")

	case ApprovalSignature:

		signature_buffer.WriteString("   /TransformMethod /FieldMDP\n")

		signature_buffer.WriteString("   /TransformParams <<\n")

		signature_buffer.WriteString("     /Type /TransformParams\n")

		signature_buffer.WriteString("     /Action /All\n")

		signature_buffer.WriteString("     /V /1.2\n")
	}

	switch context.SignData.DigestAlgorithm {
	case crypto.MD5:
		signature_buffer.WriteString("   /DigestMethod /MD5\n")
	case crypto.SHA1:
		signature_buffer.WriteString("   /DigestMethod /SHA1\n")
	case crypto.SHA256:
		signature_buffer.WriteString("   /DigestMethod /SHA256\n")
	case crypto.SHA384:
		signature_buffer.WriteString("   /DigestMethod /SHA384\n")
	case crypto.SHA512:
		signature_buffer.WriteString("   /DigestMethod /SHA512\n")
	case crypto.RIPEMD160:
		signature_buffer.WriteString("   /DigestMethod /RIPEMD160\n")
	}

	switch context.SignData.Signature.CertType {
	case CertificationSignature, UsageRightsSignature:
		signature_buffer.WriteString("   >>\n")
		signature_buffer.WriteString(" >>")
		signature_buffer.WriteString(" ]")
	}

	switch context.SignData.Signature.CertType {
	case ApprovalSignature:
		signature_buffer.WriteString(" >>\n")
	}

	if context.SignData.Signature.Info.Name != "" {
		signature_buffer.WriteString(" /Name ")
		signature_buffer.WriteString(pdfString(context.SignData.Signature.Info.Name))
		signature_buffer.WriteString("\n")
	}
	if context.SignData.Signature.Info.Location != "" {
		signature_buffer.WriteString(" /Location ")
		signature_buffer.WriteString(pdfString(context.SignData.Signature.Info.Location))
		signature_buffer.WriteString("\n")
	}
	if context.SignData.Signature.Info.Reason != "" {
		signature_buffer.WriteString(" /Reason ")
		signature_buffer.WriteString(pdfString(context.SignData.Signature.Info.Reason))
		signature_buffer.WriteString("\n")
	}
	if context.SignData.Signature.Info.ContactInfo != "" {
		signature_buffer.WriteString(" /ContactInfo ")
		signature_buffer.WriteString(pdfString(context.SignData.Signature.Info.ContactInfo))
		signature_buffer.WriteString("\n")
	}

	if context.SignData.TSA.URL == "" && !context.SignData.Signature.Info.Date.IsZero() {
		signature_buffer.WriteString(" /M ")
		signature_buffer.WriteString(pdfDateTime(context.SignData.Signature.Info.Date))
		signature_buffer.WriteString("\n")
	}

	signature_buffer.WriteString(">>\n")

	return signature_buffer.Bytes()
}

func (context *SignContext) createTimestampPlaceholder() []byte {
	var timestamp_buffer bytes.Buffer

	timestamp_buffer.WriteString("<<\n")
	timestamp_buffer.WriteString(" /Type /DocTimeStamp\n")
	timestamp_buffer.WriteString(" /Filter /Adobe.PPKLite\n")
	timestamp_buffer.WriteString(" /SubFilter /ETSI.RFC3161\n")

	timestamp_buffer.WriteString(context.createPropBuild())

	timestamp_buffer.WriteString(" " + signatureByteRangePlaceholder)

	timestamp_buffer.WriteString(" /Contents<")
	timestamp_buffer.Write(bytes.Repeat([]byte("0"), int(context.SignatureMaxLength)))
	timestamp_buffer.WriteString(">\n")
	timestamp_buffer.WriteString(">>\n")

	return timestamp_buffer.Bytes()
}

func (context *SignContext) createSignature() ([]byte, error) {
	if _, err := context.OutputBuffer.Seek(0, 0); err != nil {
		return nil, err
	}

	file_content := context.OutputBuffer.Buff.Bytes()

	sign_content := make([]byte, 0)
	sign_content = append(sign_content, file_content[context.ByteRangeValues[0]:(context.ByteRangeValues[0]+context.ByteRangeValues[1])]...)
	sign_content = append(sign_content, file_content[context.ByteRangeValues[2]:(context.ByteRangeValues[2]+context.ByteRangeValues[3])]...)

	if context.SignData.Signature.CertType == TimeStampSignature {

		timestamp_response, err := context.GetTSA(sign_content)
		if err != nil {
			return nil, fmt.Errorf("get timestamp: %w", err)
		}

		ts, err := timestamp.ParseResponse(timestamp_response)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp: %w", err)
		}

		return ts.RawToken, nil
	}

	signed_data, err := pkcs7.NewSignedData(sign_content)
	if err != nil {
		return nil, fmt.Errorf("new signed data: %w", err)
	}

	signed_data.SetDigestAlgorithm(getOIDFromHashAlgorithm(context.SignData.DigestAlgorithm))
	signingCertificate, err := context.createSigningCertificateAttribute()
	if err != nil {
		return nil, fmt.Errorf("new signed data: %w", err)
	}

	signer_config := pkcs7.SignerInfoConfig{
		ExtraSignedAttributes: []pkcs7.Attribute{
			{
				Type:  asn1.ObjectIdentifier{1, 2, 840, 113583, 1, 1, 8},
				Value: context.SignData.RevocationData,
			},
			*signingCertificate,
		},
	}

	var certificate_chain []*x509.Certificate
	if len(context.SignData.CertificateChains) > 0 && len(context.SignData.CertificateChains[0]) > 1 {
		certificate_chain = context.SignData.CertificateChains[0][1:]
	}

	if err := signed_data.AddSignerChain(context.SignData.Certificate, context.SignData.Signer, certificate_chain, signer_config); err != nil {
		return nil, fmt.Errorf("add signer chain: %w", err)
	}

	signed_data.Detach()

	if context.SignData.TSA.URL != "" {
		signature_data := signed_data.GetSignedData()

		timestamp_response, err := context.GetTSA(signature_data.SignerInfos[0].EncryptedDigest)
		if err != nil {
			return nil, fmt.Errorf("get timestamp: %w", err)
		}

		ts, err := timestamp.ParseResponse(timestamp_response)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp: %w", err)
		}

		_, err = pkcs7.Parse(ts.RawToken)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp token: %w", err)
		}

		timestamp_attribute := pkcs7.Attribute{
			Type:  asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14},
			Value: asn1.RawValue{FullBytes: ts.RawToken},
		}
		if err := signature_data.SignerInfos[0].SetUnauthenticatedAttributes([]pkcs7.Attribute{timestamp_attribute}); err != nil {
			return nil, err
		}
	}

	return signed_data.Finish()
}

func (context *SignContext) createSigningCertificateAttribute() (*pkcs7.Attribute, error) {
	hash := context.SignData.DigestAlgorithm.New()
	hash.Write(context.SignData.Certificate.Raw)

	var b cryptobyte.Builder
	b.AddASN1(cryptobyte_asn1.SEQUENCE, func(b *cryptobyte.Builder) {
		b.AddASN1(cryptobyte_asn1.SEQUENCE, func(b *cryptobyte.Builder) {
			b.AddASN1(cryptobyte_asn1.SEQUENCE, func(b *cryptobyte.Builder) {
				if context.SignData.DigestAlgorithm.HashFunc() != crypto.SHA1 &&
					context.SignData.DigestAlgorithm.HashFunc() != crypto.SHA256 {
					b.AddASN1(cryptobyte_asn1.SEQUENCE, func(b *cryptobyte.Builder) {
						b.AddASN1ObjectIdentifier(getOIDFromHashAlgorithm(context.SignData.DigestAlgorithm))
					})
				}
				b.AddASN1OctetString(hash.Sum(nil))
			})
		})
	})

	sse, err := b.Bytes()
	if err != nil {
		return nil, err
	}
	signingCertificate := pkcs7.Attribute{
		Type:  asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 47},
		Value: asn1.RawValue{FullBytes: sse},
	}
	if context.SignData.DigestAlgorithm.HashFunc() == crypto.SHA1 {
		signingCertificate.Type = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 12}
	}
	return &signingCertificate, nil
}

func (context *SignContext) updateByteRange() error {
	if _, err := context.OutputBuffer.Seek(0, 0); err != nil {
		return err
	}

	contentsPlaceholder := bytes.Repeat([]byte("0"), int(context.SignatureMaxLength))
	contentsIndex := bytes.Index(context.OutputBuffer.Buff.Bytes(), contentsPlaceholder)
	if contentsIndex == -1 {
		return fmt.Errorf("failed to find contents placeholder")
	}

	signatureContentsStart := int64(contentsIndex) - 1
	signatureContentsEnd := signatureContentsStart + int64(context.SignatureMaxLength) + 2
	context.ByteRangeValues = []int64{
		0,
		signatureContentsStart,
		signatureContentsEnd,
		int64(context.OutputBuffer.Buff.Len()) - signatureContentsEnd,
	}

	new_byte_range := fmt.Sprintf("/ByteRange [%d %d %d %d]", context.ByteRangeValues[0], context.ByteRangeValues[1], context.ByteRangeValues[2], context.ByteRangeValues[3])

	if len(new_byte_range) < len(signatureByteRangePlaceholder) {
		new_byte_range += strings.Repeat(" ", len(signatureByteRangePlaceholder)-len(new_byte_range))
	} else if len(new_byte_range) != len(signatureByteRangePlaceholder) {
		return fmt.Errorf("new byte range string is the same lenght as the placeholder")
	}

	placeholderIndex := bytes.Index(context.OutputBuffer.Buff.Bytes(), []byte(signatureByteRangePlaceholder))
	if placeholderIndex == -1 {
		return fmt.Errorf("failed to find ByteRange placeholder")
	}

	bufferBytes := context.OutputBuffer.Buff.Bytes()
	copy(bufferBytes[placeholderIndex:placeholderIndex+len(new_byte_range)], []byte(new_byte_range))

	context.OutputBuffer.Buff.Reset()
	if _, err := context.OutputBuffer.Buff.Write(bufferBytes); err != nil {
		return err
	}

	return nil
}

func (context *SignContext) replaceSignature() error {
	signature, err := context.createSignature()
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	dst := make([]byte, hex.EncodedLen(len(signature)))
	hex.Encode(dst, signature)

	if uint32(len(dst)) > context.SignatureMaxLength {
		log.Println("Signature too long, retrying with increased buffer size.")

		context.SignatureMaxLengthBase += (uint32(len(dst)) - context.SignatureMaxLength) + 1
		return context.SignPDF()
	}

	if _, err := context.OutputBuffer.Seek(0, 0); err != nil {
		return err
	}
	file_content := context.OutputBuffer.Buff.Bytes()

	if _, err := context.OutputBuffer.Write(file_content[context.ByteRangeValues[0]:context.ByteRangeValues[1]]); err != nil {
		return err
	}

	if _, err := context.OutputBuffer.Write([]byte("<")); err != nil {
		return err
	}

	if _, err := context.OutputBuffer.Write([]byte(dst)); err != nil {
		return err
	}

	zeroPadding := bytes.Repeat([]byte("0"), int(context.SignatureMaxLength)-len(dst))
	if _, err := context.OutputBuffer.Write(zeroPadding); err != nil {
		return err
	}

	if _, err := context.OutputBuffer.Write([]byte(">")); err != nil {
		return err
	}

	if _, err := context.OutputBuffer.Write(file_content[context.ByteRangeValues[2] : context.ByteRangeValues[2]+context.ByteRangeValues[3]]); err != nil {
		return err
	}

	return nil
}

func getOIDFromHashAlgorithm(target crypto.Hash) asn1.ObjectIdentifier {
	for hash, oid := range hashOIDs {
		if hash == target {
			return oid
		}
	}
	return nil
}
