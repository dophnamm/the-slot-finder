package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostJSON_errorStatus(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		wantRejected bool
	}{
		{name: "tool call rejected", status: http.StatusBadRequest, wantRejected: true,
			body: `{"error":{"message":"Tool call validation failed","type":"invalid_request_error","code":"tool_use_failed"}}`},
		{name: "other bad request", status: http.StatusBadRequest,
			body: `{"error":{"message":"bad model","code":"model_not_found"}}`},
		{name: "server error with the same code", status: http.StatusInternalServerError,
			body: `{"error":{"code":"tool_use_failed"}}`},
		{name: "body is not JSON", status: http.StatusBadRequest, body: "oops"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			err := postJSON(context.Background(), srv.Client(), srv.URL, "Authorization", "Bearer k", map[string]any{}, &struct{}{})
			require.Error(t, err)
			var rejected *ToolCallRejectedError
			assert.Equal(t, tt.wantRejected, errors.As(err, &rejected))
		})
	}
}
