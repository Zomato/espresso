package signer

import (
	"crypto"
	"encoding/asn1"
	"strconv"
)

// PDF annotation flags
const (
	AnnotationFlagInvisible      = 1 << 0
	AnnotationFlagHidden         = 1 << 1
	AnnotationFlagPrint          = 1 << 2
	AnnotationFlagNoZoom         = 1 << 3
	AnnotationFlagNoRotate       = 1 << 4
	AnnotationFlagNoView         = 1 << 5
	AnnotationFlagReadOnly       = 1 << 6
	AnnotationFlagLocked         = 1 << 7
	AnnotationFlagToggleNoView   = 1 << 8
	AnnotationFlagLockedContents = 1 << 9
)

// XREF stream constants
const (
	xrefStreamColumns   = 6
	xrefStreamPredictor = 12
	defaultPredictor    = 1
	pngSubPredictor     = 11
	pngUpPredictor      = 12
)

// Object constants
const (
	objectFooter = "\nendobj\n"
)

// Signature placeholder strings
const (
	signatureByteRangePlaceholder = "/ByteRange[0 ********** ********** **********]"
)

const (
	CertificationSignature CertType = iota + 1
	ApprovalSignature
	UsageRightsSignature
	TimeStampSignature
)

// String method for CertType
func (i CertType) String() string {
	const _CertType_name = "CertificationSignatureApprovalSignatureUsageRightsSignatureTimeStampSignature"
	var _CertType_index = [...]uint8{0, 22, 39, 59, 77}

	i -= 1
	if i >= CertType(len(_CertType_index)-1) {
		return "CertType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _CertType_name[_CertType_index[i]:_CertType_index[i+1]]
}

const (
	DoNotAllowAnyChangesPerms DocMDPPerm = iota + 1
	AllowFillingExistingFormFieldsAndSignaturesPerms
	AllowFillingExistingFormFieldsAndSignaturesAndCRUDAnnotationsPerms
)

// String method for DocMDPPerm
func (i DocMDPPerm) String() string {
	const _DocMDPPerm_name = "DoNotAllowAnyChangesPermsAllowFillingExistingFormFieldsAndSignaturesPermsAllowFillingExistingFormFieldsAndSignaturesAndCRUDAnnotationsPerms"
	var _DocMDPPerm_index = [...]uint8{0, 25, 73, 139}

	i -= 1
	if i >= DocMDPPerm(len(_DocMDPPerm_index)-1) {
		return "DocMDPPerm(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _DocMDPPerm_name[_DocMDPPerm_index[i]:_DocMDPPerm_index[i+1]]
}

var hashOIDs = map[crypto.Hash]asn1.ObjectIdentifier{
	crypto.SHA1:   asn1.ObjectIdentifier([]int{1, 3, 14, 3, 2, 26}),
	crypto.SHA256: asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 1}),
	crypto.SHA384: asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 2}),
	crypto.SHA512: asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 3}),
}
