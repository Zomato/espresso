package signer

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/digitorus/pdf"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func pdfString(text string) string {
	if !isASCII(text) {

		enc := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder()
		res, _, err := transform.String(enc, text)
		if err != nil {
			panic(err)
		}
		return "(" + res + ")"
	}

	text = strings.Replace(text, "\\", "\\\\", -1)
	text = strings.Replace(text, ")", "\\)", -1)
	text = strings.Replace(text, "(", "\\(", -1)
	text = strings.Replace(text, "\r", "\\r", -1)
	text = "(" + text + ")"

	return text
}

func pdfDateTime(date time.Time) string {

	_, original_offset := date.Zone()
	offset := original_offset
	if offset < 0 {
		offset = -offset
	}

	offset_duration := time.Duration(offset) * time.Second
	offset_hours := int(math.Floor(offset_duration.Hours()))
	offset_minutes := int(math.Floor(offset_duration.Minutes()))
	offset_minutes = offset_minutes - (offset_hours * 60)

	dateString := "D:" + date.Format("20060102150405")

	if original_offset < 0 {
		dateString += "-"
	} else {
		dateString += "+"
	}

	offset_hours_formatted := fmt.Sprintf("%d", offset_hours)
	offset_minutes_formatted := fmt.Sprintf("%d", offset_minutes)
	dateString += leftPad(offset_hours_formatted, "0", 2-len(offset_hours_formatted)) + "'" + leftPad(offset_minutes_formatted, "0", 2-len(offset_minutes_formatted)) + "'"

	return pdfString(dateString)
}

func leftPad(s string, padStr string, pLen int) string {
	if pLen <= 0 {
		return s
	}
	return strings.Repeat(padStr, pLen) + s
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > '\u007F' {
			return false
		}
	}
	return true
}

func findPageByNumber(pages pdf.Value, targetPageNumber uint32) (pdf.Value, error) {
	// Use a stack to track pages to process
	type stackItem struct {
		page           pdf.Value
		processedCount uint32
	}

	stack := []stackItem{{page: pages, processedCount: 0}}
	currentCount := uint32(0)

	for len(stack) > 0 {
		// Pop the last item from the stack
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Process based on type
		if item.page.Key("Type").Name() == "Pages" {
			kids := item.page.Key("Kids")
			for i := kids.Len() - 1; i >= 0; i-- {
				// Push kids to the stack in reverse order to maintain order when popping
				stack = append(stack, stackItem{
					page:           kids.Index(i),
					processedCount: item.processedCount,
				})
			}
		} else if item.page.Key("Type").Name() == "Page" {
			currentCount++
			if currentCount == targetPageNumber {
				return item.page, nil
			}
		}
	}

	return pdf.Value{}, fmt.Errorf("page number %d not found", targetPageNumber)
}
