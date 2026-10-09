package dtos

import (
	"testing"

	"the-slot-finder/server/internal/constants"

	"github.com/stretchr/testify/require"
)

func TestPageableRequest_Normalize(t *testing.T) {
	tests := []struct {
		name         string
		in           PageableRequest
		wantPage     int
		wantPageSize int
	}{
		{name: "defaults when empty", in: PageableRequest{}, wantPage: constants.DefaultPage, wantPageSize: constants.DefaultPageSize},
		{name: "keeps valid values", in: PageableRequest{Page: 3, PageSize: 25}, wantPage: 3, wantPageSize: 25},
		{name: "caps page size", in: PageableRequest{Page: 1, PageSize: constants.MaxPageSize + 1}, wantPage: 1, wantPageSize: constants.MaxPageSize},
		{name: "fixes negatives", in: PageableRequest{Page: -1, PageSize: -5}, wantPage: constants.DefaultPage, wantPageSize: constants.DefaultPageSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := tt.in
			pr.Normalize()
			require.Equal(t, tt.wantPage, pr.Page)
			require.Equal(t, tt.wantPageSize, pr.PageSize)
		})
	}
}
