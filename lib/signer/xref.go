package signer

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (context *SignContext) writeXref() error {
	if _, err := context.OutputBuffer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline before xref: %w", err)
	}
	context.NewXrefStart = int64(context.OutputBuffer.Buff.Len())

	switch context.PDFReader.XrefInformation.Type {
	case "table":
		return context.writeIncrXrefTable()
	case "stream":
		return context.writeXrefStream()
	default:
		return fmt.Errorf("unknown xref type: %s", context.PDFReader.XrefInformation.Type)
	}
}

func (context *SignContext) getLastObjectIDFromXref() (uint32, error) {
	xref := context.PDFReader.Xref()
	if len(xref) == 0 {
		return 0, fmt.Errorf("no xref entries found")
	}

	var maxID uint32
	for _, entry := range xref {
		ptr := entry.Ptr()

		if ptr.GetID() > maxID {
			maxID = ptr.GetID()
		}
	}

	return maxID + 1, nil
}

func (context *SignContext) writeIncrXrefTable() error {

	if _, err := context.OutputBuffer.Write([]byte("xref\n")); err != nil {
		return fmt.Errorf("failed to write incremental xref header: %w", err)
	}

	for _, entry := range context.updatedXrefEntries {
		pageXrefObj := fmt.Sprintf("%d %d\n", entry.ID, 1)
		if _, err := context.OutputBuffer.Write([]byte(pageXrefObj)); err != nil {
			return fmt.Errorf("failed to write updated xref object: %w", err)
		}

		xrefLine := fmt.Sprintf("%010d 00000 n\r\n", entry.Offset)
		if _, err := context.OutputBuffer.Write([]byte(xrefLine)); err != nil {
			return fmt.Errorf("failed to write updated incremental xref entry: %w", err)
		}
	}

	startXrefObj := fmt.Sprintf("%d %d\n", context.lastXrefID+1, len(context.newXrefEntries))
	if _, err := context.OutputBuffer.Write([]byte(startXrefObj)); err != nil {
		return fmt.Errorf("failed to write starting xref object: %w", err)
	}

	for _, entry := range context.newXrefEntries {
		xrefLine := fmt.Sprintf("%010d 00000 n\r\n", entry.Offset)
		if _, err := context.OutputBuffer.Write([]byte(xrefLine)); err != nil {
			return fmt.Errorf("failed to write incremental xref entry: %w", err)
		}
	}

	return nil
}

func (context *SignContext) writeXrefStream() error {
	var buffer bytes.Buffer

	predictor := context.PDFReader.Trailer().Key("DecodeParms").Key("Predictor").Int64()
	if predictor == 0 {
		predictor = xrefStreamPredictor
	}

	if err := writeXrefStreamEntries(&buffer, context); err != nil {
		return fmt.Errorf("failed to write xref stream entries: %w", err)
	}

	streamBytes, err := encodeXrefStream(buffer.Bytes(), predictor)
	if err != nil {
		return fmt.Errorf("failed to encode xref stream: %w", err)
	}

	var xrefStreamObject bytes.Buffer

	if err := writeXrefStreamHeader(&xrefStreamObject, context, len(streamBytes)); err != nil {
		return fmt.Errorf("failed to write xref stream header: %w", err)
	}

	if err := writeXrefStreamContent(&xrefStreamObject, streamBytes); err != nil {
		return fmt.Errorf("failed to write xref stream content: %w", err)
	}

	_, err = context.addObject(xrefStreamObject.Bytes())
	if err != nil {
		return fmt.Errorf("failed to add xref stream object: %w", err)
	}

	return nil
}

func writeXrefStreamEntries(buffer *bytes.Buffer, context *SignContext) error {

	for _, entry := range context.updatedXrefEntries {
		writeXrefStreamLine(buffer, 1, int(entry.Offset), 0)
	}

	for _, entry := range context.newXrefEntries {
		writeXrefStreamLine(buffer, 1, int(entry.Offset), 0)
	}

	return nil
}

func encodeXrefStream(data []byte, predictor int64) ([]byte, error) {

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

func writeXrefStreamHeader(buffer *bytes.Buffer, context *SignContext, streamLength int) error {
	id := context.PDFReader.Trailer().Key("ID")

	totalEntries := uint32(context.PDFReader.XrefInformation.ItemCount)
	var indexArray []uint32

	if len(context.updatedXrefEntries) > 0 {
		for _, entry := range context.updatedXrefEntries {
			indexArray = append(indexArray, entry.ID, 1)
		}
	}

	if len(context.newXrefEntries) > 0 {
		indexArray = append(indexArray, context.lastXrefID+1, uint32(len(context.newXrefEntries)))
		totalEntries += uint32(len(context.newXrefEntries))
	}

	buffer.WriteString("<< /Type /XRef\n")
	buffer.WriteString(fmt.Sprintf("  /Length %d\n", streamLength))
	buffer.WriteString("  /Filter /FlateDecode\n")

	buffer.WriteString("  /W [ 1 4 1 ]\n")
	buffer.WriteString(fmt.Sprintf("  /Prev %d\n", context.PDFReader.XrefInformation.StartPos))
	buffer.WriteString(fmt.Sprintf("  /Size %d\n", totalEntries+1))

	if len(indexArray) > 0 {
		buffer.WriteString("  /Index [")
		for _, idx := range indexArray {
			buffer.WriteString(fmt.Sprintf(" %d", idx))
		}
		buffer.WriteString(" ]\n")
	}

	buffer.WriteString(fmt.Sprintf("  /Root %d 0 R\n", context.CatalogData.ObjectId))

	if !id.IsNull() {
		id0 := hex.EncodeToString([]byte(id.Index(0).RawString()))
		id1 := hex.EncodeToString([]byte(id.Index(1).RawString()))
		buffer.WriteString(fmt.Sprintf("  /ID [<%s><%s>]\n", id0, id1))
	}

	buffer.WriteString(">>\n")
	return nil
}

func writeXrefStreamContent(buffer *bytes.Buffer, streamBytes []byte) error {
	if _, err := io.WriteString(buffer, "stream\n"); err != nil {
		return err
	}

	if _, err := buffer.Write(streamBytes); err != nil {
		return err
	}

	if _, err := io.WriteString(buffer, "\nendstream\n"); err != nil {
		return err
	}

	return nil
}

func writeXrefStreamLine(b *bytes.Buffer, xreftype byte, offset int, gen byte) {

	b.WriteByte(xreftype)

	offsetBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(offsetBytes, uint32(offset))
	b.Write(offsetBytes)

	b.WriteByte(gen)
}

func (context *SignContext) writeTrailer() error {
	if context.PDFReader.XrefInformation.Type == "table" {
		trailer_length := context.PDFReader.XrefInformation.IncludingTrailerEndPos - context.PDFReader.XrefInformation.EndPos

		if _, err := context.InputFile.Seek(context.PDFReader.XrefInformation.EndPos+1, 0); err != nil {
			return err
		}
		trailer_buf := make([]byte, trailer_length)
		if _, err := context.InputFile.Read(trailer_buf); err != nil {
			return err
		}

		root_string := "Root " + context.CatalogData.RootString
		new_root := "Root " + strconv.FormatInt(int64(context.CatalogData.ObjectId), 10) + " 0 R"

		size_string := "Size " + strconv.FormatInt(context.PDFReader.XrefInformation.ItemCount, 10)
		new_size := "Size " + strconv.FormatInt(context.PDFReader.XrefInformation.ItemCount+int64(len(context.newXrefEntries)+1), 10)

		prev_string := "Prev " + context.PDFReader.Trailer().Key("Prev").String()
		new_prev := "Prev " + strconv.FormatInt(context.PDFReader.XrefInformation.StartPos, 10)

		trailer_string := string(trailer_buf)
		trailer_string = strings.Replace(trailer_string, root_string, new_root, -1)
		trailer_string = strings.Replace(trailer_string, size_string, new_size, -1)
		if strings.Contains(trailer_string, prev_string) {
			trailer_string = strings.Replace(trailer_string, prev_string, new_prev, -1)
		} else {
			trailer_string = strings.Replace(trailer_string, new_root, new_root+"\n  /"+new_prev, -1)
		}

		lines := strings.Split(trailer_string, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, " ") {
				lines[i] = "    " + strings.TrimSpace(line)
			}
		}
		trailer_string = strings.Join(lines, "\n")

		if _, err := context.OutputBuffer.Write([]byte(trailer_string)); err != nil {
			return err
		}
	} else if context.PDFReader.XrefInformation.Type == "stream" {
		if _, err := context.OutputBuffer.Write([]byte("startxref\n")); err != nil {
			return err
		}
	}

	if _, err := context.OutputBuffer.Write([]byte(strconv.FormatInt(context.NewXrefStart, 10) + "\n")); err != nil {
		return err
	}

	if _, err := context.OutputBuffer.Write([]byte("%%EOF\n")); err != nil {
		return err
	}

	return nil
}
