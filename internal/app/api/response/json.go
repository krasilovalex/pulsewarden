package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(
	w http.ResponseWriter,
	status int,
	payload any,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("encode JSON response: %w", err)
	}

	return nil
}
