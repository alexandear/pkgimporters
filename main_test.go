package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestFetchImporterCount(t *testing.T) {
	tests := []struct {
		name          string
		htmlFile      string
		pkgPath       string
		expectedCount int
		expectedURL   string
	}{
		{
			name:          "io package",
			htmlFile:      "testdata/io.html",
			pkgPath:       "io",
			expectedCount: 1533321,
			expectedURL:   "https://pkg.go.dev/io?tab=importedby",
		},
		{
			name:          "golang.org/x/tools/go/analysis package",
			htmlFile:      "testdata/golang.org/x/tools/go/analysis.html",
			pkgPath:       "golang.org/x/tools/go/analysis",
			expectedCount: 6136,
			expectedURL:   "https://pkg.go.dev/golang.org/x/tools/go/analysis?tab=importedby",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlBytes, err := os.ReadFile(tt.htmlFile)
			if err != nil {
				t.Fatal(err)
			}

			transport := &htmlFileTransport{
				content: htmlBytes,
			}
			client := &http.Client{
				Transport: transport,
			}
			count, err := fetchImporterCount(t.Context(), client, tt.pkgPath)
			if err != nil {
				t.Fatal(err)
			}

			if count != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, count)
			}
			if len(transport.requestedURLs) != 1 {
				t.Fatalf("expected 1 request, got %d", len(transport.requestedURLs))
			}
			if transport.requestedURLs[0] != tt.expectedURL {
				t.Errorf("expected URL %q, got %q", tt.expectedURL, transport.requestedURLs[0])
			}
		})
	}
}

// htmlFileTransport is a custom http.RoundTripper that returns test HTML content
type htmlFileTransport struct {
	content       []byte
	requestedURLs []string
}

func (t *htmlFileTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.requestedURLs = append(t.requestedURLs, req.URL.String())
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Header: http.Header{
			"Content-Type": []string{"text/html"},
		},
		Body:          io.NopCloser(bytes.NewReader(t.content)),
		ContentLength: int64(len(t.content)),
		Request:       req,
	}, nil
}
