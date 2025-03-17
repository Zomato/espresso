package signer

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/digitorus/pdf"
)

func (context *SignContext) createCatalog() ([]byte, error) {
	var catalog_buffer bytes.Buffer

	catalog_buffer.WriteString("<<\n")
	catalog_buffer.WriteString("  /Type /Catalog\n")

	root := context.PDFReader.Trailer().Key("Root")
	rootPtr := root.GetPtr()
	context.CatalogData.RootString = strconv.Itoa(int(rootPtr.GetID())) + " " + strconv.Itoa(int(rootPtr.GetGen())) + " R"

	for _, key := range root.Keys() {
		if key != "Type" && key != "AcroForm" {
			_, _ = fmt.Fprintf(&catalog_buffer, "  /%s ", key)
			context.serializeCatalogEntry(&catalog_buffer, rootPtr.GetID(), root.Key(key))
			catalog_buffer.WriteString("\n")
		}
	}

	catalog_buffer.WriteString("  /AcroForm <<\n")
	catalog_buffer.WriteString("    /Fields [")

	for i, sig := range context.existingSignatures {
		if i > 0 {
			catalog_buffer.WriteString(" ")
		}
		catalog_buffer.WriteString(strconv.Itoa(int(sig.objectId)) + " 0 R")
	}

	if len(context.existingSignatures) > 0 {
		catalog_buffer.WriteString(" ")
	}
	catalog_buffer.WriteString(strconv.Itoa(int(context.VisualSignData.objectId)) + " 0 R")

	catalog_buffer.WriteString("]\n")

	switch context.SignData.Signature.CertType {
	case CertificationSignature, ApprovalSignature, TimeStampSignature:
		catalog_buffer.WriteString("    /SigFlags 3\n")
	case UsageRightsSignature:
		catalog_buffer.WriteString("    /SigFlags 1\n")
	}

	catalog_buffer.WriteString("  >>\n")
	catalog_buffer.WriteString(">>\n")

	return catalog_buffer.Bytes(), nil
}

func (context *SignContext) serializeCatalogEntry(w io.Writer, rootObjId uint32, value pdf.Value) {
	// Define a stack item type to track our work
	type stackItem struct {
		val   pdf.Value
		state int      // 0: new, 1: started, 2: processing array/dict items
		index int      // current index for arrays/dicts
		keys  []string // for dictionaries
	}

	// Initialize the stack with our first value
	stack := []stackItem{{val: value, state: 0}}

	for len(stack) > 0 {
		// Get the current item (without popping)
		idx := len(stack) - 1
		current := &stack[idx]

		// Handle according to state
		switch current.state {
		case 0: // New item, determine type and start processing
			if ptr := current.val.GetPtr(); ptr.GetID() != rootObjId {
				// Direct reference handling
				fmt.Fprintf(w, "%d %d R", ptr.GetID(), ptr.GetGen())
				stack = stack[:idx] // pop
				continue
			}

			// Process based on type
			switch current.val.Kind() {
			case pdf.String:
				fmt.Fprintf(w, "(%s)", current.val.RawString())
				stack = stack[:idx] // pop

			case pdf.Null:
				fmt.Fprint(w, "null")
				stack = stack[:idx] // pop

			case pdf.Bool:
				if current.val.Bool() {
					fmt.Fprint(w, "true")
				} else {
					fmt.Fprint(w, "false")
				}
				stack = stack[:idx] // pop

			case pdf.Integer:
				fmt.Fprintf(w, "%d", current.val.Int64())
				stack = stack[:idx] // pop

			case pdf.Real:
				fmt.Fprintf(w, "%f", current.val.Float64())
				stack = stack[:idx] // pop

			case pdf.Name:
				fmt.Fprintf(w, "/%s", current.val.Name())
				stack = stack[:idx] // pop

			case pdf.Dict:
				fmt.Fprint(w, "<<")
				current.keys = current.val.Keys()
				current.index = 0
				current.state = 2 // Processing dict items

			case pdf.Array:
				fmt.Fprint(w, "[")
				current.index = 0
				current.state = 2 // Processing array items

			case pdf.Stream:
				panic("stream cannot be a direct object")
			}

		case 2: // Processing container items
			if current.val.Kind() == pdf.Dict {
				// Dictionary processing
				if current.index >= len(current.keys) {
					fmt.Fprint(w, ">>")
					stack = stack[:idx] // pop
				} else {
					if current.index > 0 {
						fmt.Fprint(w, " ")
					}
					key := current.keys[current.index]
					fmt.Fprintf(w, "/%s ", key)
					// Push the value onto the stack
					stack = append(stack, stackItem{val: current.val.Key(key), state: 0})
					current.index++
				}
			} else if current.val.Kind() == pdf.Array {
				// Array processing
				if current.index >= current.val.Len() {
					fmt.Fprint(w, "]")
					stack = stack[:idx] // pop
				} else {
					if current.index > 0 {
						fmt.Fprint(w, " ")
					}
					// Push the value onto the stack
					stack = append(stack, stackItem{val: current.val.Index(current.index), state: 0})
					current.index++
				}
			}
		}
	}
}
