package signer

import (
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"io"
	"time"

	"github.com/digitorus/pdf"
	"github.com/mattetti/filebuffer"
)

// CatalogData holds information about the PDF catalog
type CatalogData struct {
	ObjectId   uint32
	RootString string
}

// TSA holds timestamp authority configuration
type TSA struct {
	URL      string
	Username string
	Password string
}

type CRL []asn1.RawValue

type OCSP []asn1.RawValue

type Other struct {
	Type  asn1.ObjectIdentifier
	Value []byte
}

type InfoArchival struct {
	CRL   CRL   `asn1:"tag:0,optional,explicit"`
	OCSP  OCSP  `asn1:"tag:1,optional,explicit"`
	Other Other `asn1:"tag:2,optional,explicit"`
}

// RevocationFunction defines a function type for checking revocation status
type RevocationFunction func(cert, issuer *x509.Certificate, i *InfoArchival) error

// SignData contains all information needed for signing a PDF
type SignData struct {
	Signature          SignDataSignature
	Signer             crypto.Signer
	DigestAlgorithm    crypto.Hash
	Certificate        *x509.Certificate
	CertificateChains  [][]*x509.Certificate
	TSA                TSA
	RevocationData     InfoArchival
	RevocationFunction RevocationFunction
	Appearance         Appearance

	objectId uint32
}

type CertType uint
type DocMDPPerm uint

// SignDataSignature contains signature metadata
type SignDataSignature struct {
	CertType   CertType
	DocMDPPerm DocMDPPerm
	Info       SignDataSignatureInfo
}

// SignDataSignatureInfo holds human-readable signature information
type SignDataSignatureInfo struct {
	Name        string
	Location    string
	Reason      string
	ContactInfo string
	Date        time.Time
}

// Appearance defines how and where a visible signature appears
type Appearance struct {
	Visible     bool
	Page        uint32
	LowerLeftX  float64
	LowerLeftY  float64
	UpperRightX float64
	UpperRightY float64
}

// VisualSignData contains object IDs for the visual signature
type VisualSignData struct {
	pageObjectId uint32
	objectId     uint32
}

// InfoData contains metadata for the signature
type InfoData struct {
	ObjectId uint32
}

// SignContext maintains the state during the signing process
type SignContext struct {
	InputFile              io.ReadSeeker
	OutputFile             io.Writer
	OutputBuffer           *filebuffer.Buffer
	SignData               SignData
	CatalogData            CatalogData
	VisualSignData         VisualSignData
	InfoData               InfoData
	PDFReader              *pdf.Reader
	NewXrefStart           int64
	ByteRangeValues        []int64
	SignatureMaxLength     uint32
	SignatureMaxLengthBase uint32

	existingSignatures []SignData
	lastXrefID         uint32
	newXrefEntries     []xrefEntry
	updatedXrefEntries []xrefEntry
}

// xrefEntry represents an entry in the PDF cross-reference table
type xrefEntry struct {
	ID         uint32
	Offset     int64
	Generation int
	Free       bool
}
