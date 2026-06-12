package service

import (
	"errors"
	"testing"
)

func TestNormalizeGitHubURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantURL string
		wantErr bool
	}{
		{
			name:    "valid repository url",
			rawURL:  "https://github.com/octocat/Hello-World",
			wantURL: "https://github.com/octocat/Hello-World",
		},
		{
			name:    "normalizes host case",
			rawURL:  "https://GitHub.com/octocat/Hello-World",
			wantURL: "https://github.com/octocat/Hello-World",
		},
		{
			name:    "rejects http",
			rawURL:  "http://github.com/octocat/Hello-World",
			wantErr: true,
		},
		{
			name:    "rejects user info",
			rawURL:  "https://user@github.com/octocat/Hello-World",
			wantErr: true,
		},
		{
			name:    "rejects non github host",
			rawURL:  "https://example.com/octocat/Hello-World",
			wantErr: true,
		},
		{
			name:    "rejects empty path",
			rawURL:  "https://github.com/",
			wantErr: true,
		},
		{
			name:    "rejects fragment",
			rawURL:  "https://github.com/octocat/Hello-World#readme",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotURL, err := NormalizeGitHubURL(tt.rawURL)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidGitHubURL) {
					t.Fatalf("expected ErrInvalidGitHubURL, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotURL != tt.wantURL {
				t.Fatalf("url mismatch: got %q want %q", gotURL, tt.wantURL)
			}
		})
	}
}
