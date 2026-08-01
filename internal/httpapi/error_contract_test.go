package httpapi

import (
	"encoding/json"
	"testing"
)

func assertStableErrorBody(t *testing.T, body []byte, wantCode string) {
	t.Helper()
	var response ErrorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("error response is not JSON: %v body=%s", err, body)
	}
	if response.SchemaVersion != ErrorSchemaVersion || response.Error.Code != wantCode ||
		response.Error.Message == "" || response.Error.Details == nil || response.Error.RequestID == "" ||
		response.Error.Timestamp.IsZero() || response.Error.Documentation == "" {
		t.Fatalf("incomplete stable error response: %+v", response)
	}
}
