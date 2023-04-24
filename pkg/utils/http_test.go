package utils

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripModifierFunc(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		res    *http.Response
		header http.Header
		body   []byte
	}{
		{
			name: "modified",
			s:    "/api/v1/namespaces/default/services/nginx/proxy",
			res: &http.Response{
				Header: map[string][]string{
					"Location":       {"/api/v1/namespaces/default/services/nginx/proxy/login"},
					"Content-Length": {"86"},
				},
				Body: io.NopCloser(bytes.NewReader([]byte(`<script src="/api/v1/namespaces/default/services/nginx/proxy/public/main.js"></script>`))),
			},
			header: http.Header(map[string][]string{
				"Location":       {"/login"},
				"Content-Length": {"39"},
			}),
			body: []byte(`<script src="/public/main.js"></script>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := StripModifierFunc(tt.s)
			_ = modifier(tt.res)

			assert.Equal(t, tt.header, tt.res.Header)

			b, _ := io.ReadAll(tt.res.Body)
			assert.Equal(t, tt.body, b)
		})
	}
}
