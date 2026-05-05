package confluence

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteURLForScopedToken(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		cloudID string
		want    string
		wantErr bool
	}{
		{
			name:    "standard atlassian.net URL with /wiki",
			baseURL: "https://mysite.atlassian.net/wiki",
			cloudID: "abc123",
			want:    "https://api.atlassian.com/ex/confluence/abc123/wiki",
		},
		{
			name:    "URL with trailing slash on /wiki",
			baseURL: "https://mysite.atlassian.net/wiki/",
			cloudID: "abc123",
			want:    "https://api.atlassian.com/ex/confluence/abc123/wiki",
		},
		{
			name:    "URL without /wiki path",
			baseURL: "https://mysite.atlassian.net",
			cloudID: "abc123",
			want:    "https://api.atlassian.com/ex/confluence/abc123/wiki",
		},
		{
			name:    "UUID cloud ID",
			baseURL: "https://example.atlassian.net/wiki",
			cloudID: "11111111-2222-3333-4444-555555555555",
			want:    "https://api.atlassian.com/ex/confluence/11111111-2222-3333-4444-555555555555/wiki",
		},
		{
			name:    "URL with suffix after /wiki",
			baseURL: "https://mysite.atlassian.net/wiki/subpath",
			cloudID: "abc123",
			want:    "https://api.atlassian.com/ex/confluence/abc123/wiki/subpath",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RewriteURLForScopedToken(tt.baseURL, tt.cloudID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFetchCloudID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/_edge/tenant_info", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"cloudId":"test-cloud-id-123"}`))
		}))
		defer srv.Close()

		cloudID, err := FetchCloudID(srv.URL)
		require.NoError(t, err)
		assert.Equal(t, "test-cloud-id-123", cloudID)
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := FetchCloudID(srv.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "returned status 500")
	})

	t.Run("empty cloud ID in response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"cloudId":""}`))
		}))
		defer srv.Close()

		_, err := FetchCloudID(srv.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty cloudId")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`not json`))
		}))
		defer srv.Close()

		_, err := FetchCloudID(srv.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to decode")
	})
}
