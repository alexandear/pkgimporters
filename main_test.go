package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	binPath := filepath.Join(t.TempDir(), "pkgimporters")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	tests := []struct {
		name        string
		args        []string
		checkOutput func(t *testing.T, output string)
		checkStderr func(t *testing.T, output string)
	}{
		{
			name: "no arguments should fail",
			args: []string{},
			checkStderr: func(t *testing.T, output string) {
				if !strings.Contains(output, "no packages specified") {
					t.Errorf("stderr should mention missing packages, got:\n%s", output)
				}
			},
		},
		{
			name: "one package",
			args: []string{"fmt"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.HasPrefix(output, "fmt") {
					t.Errorf("output should start with 'fmt', got:\n%s", output)
				}
				assertCountsArePositive(t, output)
			},
		},
		{
			name: "comma-separated packages",
			args: []string{"-pkgs", "fmt,io"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "fmt") || !strings.Contains(output, "io") {
					t.Errorf("output should contain both 'fmt' and 'io', got:\n%s", output)
				}
				assertCountsArePositive(t, output)
			},
		},
		{
			name: "sort by count",
			args: []string{"-sort", "count", "fmt", "io"},
			checkOutput: func(t *testing.T, output string) {
				fmtIdx := strings.Index(output, "fmt ")
				ioIdx := strings.Index(output, "io ")
				if fmtIdx < 0 || ioIdx < 0 {
					t.Fatalf("could not find 'fmt' or 'io' in output:\n%s", output)
				}
				if fmtIdx > ioIdx {
					t.Errorf("'fmt' should appear before 'io' when sorted by count descending, got:\n%s", output)
				}
				assertCountsArePositive(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			stderrOutput := stderr.String()
			if tt.checkStderr != nil {
				if err == nil {
					t.Error("expected command to fail, but it succeeded")
					return
				}
				tt.checkStderr(t, stderrOutput)
				return
			}

			if err != nil {
				t.Fatalf("command failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}

			output := stdout.String()
			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func assertCountsArePositive(t *testing.T, output string) {
	t.Helper()

	for line := range strings.Lines(output) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			t.Fatalf("unexpected output line: %s", line)
		}

		countStr := strings.ReplaceAll(fields[len(fields)-1], ",", "")
		count, err := strconv.Atoi(countStr)
		if err != nil {
			t.Fatalf("failed to parse count %q in line: %s", fields[len(fields)-1], line)
		}

		if count <= 0 {
			t.Errorf("expected count > 0, got %d in line: %s", count, line)
		}
	}
}
