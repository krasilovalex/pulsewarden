package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	requestID string,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	payload := ErrorEnvelope{
		Error: ErrorBody{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("encode error response: %w", err)
	}

	return nil
}
