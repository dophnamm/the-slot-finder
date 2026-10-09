package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestRespondAuthError_UsesEnvelope(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantCode string
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, wantCode: constants.Unauthorized},
		{name: "forbidden", status: http.StatusForbidden, wantCode: constants.Forbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

			require.NoError(t, respondAuthError(c, tt.status, constants.MsgTokenInvalid))
			require.Equal(t, tt.status, rec.Code)

			var body dtos.BaseResponse[any]
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Equal(t, tt.wantCode, body.Meta.ErrorCode)
			require.Equal(t, constants.MsgTokenInvalid, body.Meta.Message)
		})
	}
}
