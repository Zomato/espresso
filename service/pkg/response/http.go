package response

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func RespondWithError(w http.ResponseWriter, message string, statusCode int) {
	errorResponse := map[string]interface{}{
		"status": map[string]string{
			"status":  "failed",
			"message": message,
		},
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func RespondWithFile(w http.ResponseWriter, fileName string, contentType string, fileBytes []byte) error {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Length", strconv.Itoa(len(fileBytes)))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(fileBytes)
	return err
}
