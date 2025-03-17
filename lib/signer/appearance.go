package signer

import (
	"bytes"
	"fmt"
	"strconv"
)

func (context *SignContext) createVisualSignature(visible bool, pageNumber uint32, rect [4]float64) ([]byte, error) {
	var visual_signature bytes.Buffer

	visual_signature.WriteString("<<\n")

	visual_signature.WriteString("  /Type /Annot\n")

	visual_signature.WriteString("  /Subtype /Widget\n")

	if visible {

		visual_signature.WriteString(fmt.Sprintf("  /Rect [%f %f %f %f]\n", rect[0], rect[1], rect[2], rect[3]))

		appearance, err := context.createAppearance(rect)
		if err != nil {
			return nil, fmt.Errorf("failed to create appearance: %w", err)
		}

		appearanceObjectId, err := context.addObject(appearance)
		if err != nil {
			return nil, fmt.Errorf("failed to add appearance object: %w", err)
		}

		visual_signature.WriteString(fmt.Sprintf("  /AP << /N %d 0 R >>\n", appearanceObjectId))

	} else {

		visual_signature.WriteString("  /Rect [0 0 0 0]\n")
	}

	root := context.PDFReader.Trailer().Key("Root")

	root_keys := root.Keys()
	found_pages := false
	for _, key := range root_keys {
		if key == "Pages" {

			found_pages = true
			break
		}
	}

	rootPtr := root.GetPtr()

	context.CatalogData.RootString = strconv.Itoa(int(rootPtr.GetID())) + " " + strconv.Itoa(int(rootPtr.GetGen())) + " R"

	if found_pages {

		page, err := findPageByNumber(root.Key("Pages"), pageNumber)
		if err != nil {
			return nil, err
		}

		page_ptr := page.GetPtr()

		context.VisualSignData.pageObjectId = page_ptr.GetID()

		visual_signature.WriteString("  /P " + strconv.Itoa(int(page_ptr.GetID())) + " " + strconv.Itoa(int(page_ptr.GetGen())) + " R\n")
	}

	annotationFlags := AnnotationFlagPrint | AnnotationFlagLocked
	visual_signature.WriteString(fmt.Sprintf("  /F %d\n", annotationFlags))

	visual_signature.WriteString("  /FT /Sig\n")

	visual_signature.WriteString(fmt.Sprintf("  /T %s\n", pdfString("Signature "+strconv.Itoa(len(context.existingSignatures)+1))))

	visual_signature.WriteString(fmt.Sprintf("  /V %d 0 R\n", context.SignData.objectId))

	visual_signature.WriteString(">>\n")

	return visual_signature.Bytes(), nil
}

func (context *SignContext) createIncPageUpdate(pageNumber, annot uint32) ([]byte, error) {
	var page_buffer bytes.Buffer

	root := context.PDFReader.Trailer().Key("Root")
	page, err := findPageByNumber(root.Key("Pages"), pageNumber)
	if err != nil {
		return nil, err
	}

	page_buffer.WriteString("<<\n")

	for _, key := range page.Keys() {
		switch key {
		case "Contents", "Parent":
			ptr := page.Key(key).GetPtr()
			page_buffer.WriteString(fmt.Sprintf("  /%s %d 0 R\n", key, ptr.GetID()))
		case "Annots":
			page_buffer.WriteString("  /Annots [\n")
			for i := 0; i < page.Key("Annots").Len(); i++ {
				ptr := page.Key(key).Index(i).GetPtr()
				page_buffer.WriteString(fmt.Sprintf("    %d 0 R\n", ptr.GetID()))
			}
			page_buffer.WriteString(fmt.Sprintf("    %d 0 R\n", annot))
			page_buffer.WriteString("  ]\n")
		default:
			page_buffer.WriteString(fmt.Sprintf("  /%s %s\n", key, page.Key(key).String()))
		}
	}

	if page.Key("Annots").IsNull() {
		page_buffer.WriteString(fmt.Sprintf("  /Annots [%d 0 R]\n", annot))
	}

	page_buffer.WriteString(">>\n")

	return page_buffer.Bytes(), nil
}

func (context *SignContext) createAppearance(rect [4]float64) ([]byte, error) {
	text := context.SignData.Signature.Info.Name

	rectWidth := rect[2] - rect[0]
	rectHeight := rect[3] - rect[1]

	if rectWidth < 1 || rectHeight < 1 {
		return nil, fmt.Errorf("invalid rectangle dimensions: width %.2f and height %.2f must be greater than 0", rectWidth, rectHeight)
	}

	fontSize := rectHeight * 0.8
	textWidth := float64(len(text)) * fontSize * 0.5
	if textWidth > rectWidth {
		fontSize = rectWidth / (float64(len(text)) * 0.5)
	}

	var appearance_stream_buffer bytes.Buffer
	appearance_stream_buffer.WriteString("q\n")
	appearance_stream_buffer.WriteString("BT\n")
	appearance_stream_buffer.WriteString(fmt.Sprintf("/F1 %.2f Tf\n", fontSize))
	appearance_stream_buffer.WriteString(fmt.Sprintf("0 %.2f Td\n", rectHeight-fontSize))
	appearance_stream_buffer.WriteString("0.2 0.2 0.6 rg\n")
	appearance_stream_buffer.WriteString(fmt.Sprintf("%s Tj\n", pdfString(text)))
	appearance_stream_buffer.WriteString("ET\n")
	appearance_stream_buffer.WriteString("Q\n")

	var appearance_buffer bytes.Buffer
	appearance_buffer.WriteString("<<\n")
	appearance_buffer.WriteString("  /Type /XObject\n")
	appearance_buffer.WriteString("  /Subtype /Form\n")
	appearance_buffer.WriteString(fmt.Sprintf("  /BBox [0 0 %f %f]\n", rectWidth, rectHeight))
	appearance_buffer.WriteString("  /Matrix [1 0 0 1 0 0]\n")

	appearance_buffer.WriteString("  /Resources <<\n")
	appearance_buffer.WriteString("   /Font <<\n")
	appearance_buffer.WriteString("     /F1 <<\n")
	appearance_buffer.WriteString("       /Type /Font\n")
	appearance_buffer.WriteString("       /Subtype /Type1\n")
	appearance_buffer.WriteString("       /BaseFont /Times-Roman\n")
	appearance_buffer.WriteString("     >>\n")
	appearance_buffer.WriteString("   >>\n")
	appearance_buffer.WriteString("  >>\n")

	appearance_buffer.WriteString("  /FormType 1\n")
	appearance_buffer.WriteString(fmt.Sprintf("  /Length %d\n", appearance_stream_buffer.Len()))
	appearance_buffer.WriteString(">>\n")

	appearance_buffer.WriteString("stream\n")
	appearance_buffer.Write(appearance_stream_buffer.Bytes())
	appearance_buffer.WriteString("endstream\n")

	return appearance_buffer.Bytes(), nil
}
