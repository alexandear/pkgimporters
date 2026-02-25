// pkgimporters fetches the number of known importers for Go packages from pkg.go.dev.
// It supports multiple ways to specify packages:
// via positional arguments, comma-separated list with -pkgs, or all stdlib with -pkgs std.
//
// Usage:
//
//	pkgimporters fmt bufio net/http          # specific packages
//	pkgimporters -pkgs fmt,bufio,net/http    # comma-separated packages
//	pkgimporters std                         # all standard library packages
//	pkgimporters -pkgs std -sort count       # sort by importer count descending
//	pkgimporters -workers 10 -pkgs std       # with tuned concurrency
package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"golang.org/x/tools/go/packages"
)

type pkgImporter struct {
	path  string
	count int
}

type cmdError struct {
	code int
	msg  string
}

func (e *cmdError) Error() string {
	return e.msg
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var e *cmdError
		if errors.As(err, &e) {
			os.Exit(e.code)
		}
		os.Exit(1)
	}
}

func run() error {
	workers := flag.Int("workers", 5, "number of concurrent requests")
	sortBy := flag.String("sort", "name", "sort results by 'name' (default) or 'count' (descending)")
	pkgsList := flag.String("pkgs", "", "comma-separated list of packages to fetch or 'std' for all standard library packages")
	progName := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "NAME\n"+
			"    %[1]s - fetch known importers for Go packages from pkg.go.dev\n\n"+
			"SYNOPSIS\n"+
			"    %[1]s [-pkgs pkg1,pkg2,...|std] [-workers N] [-sort name|count] [package ...]\n\n"+
			"DESCRIPTION\n"+
			"    %[1]s fetches the number of known importers for Go packages from https://pkg.go.dev.\n"+
			"Packages can be specified via positional arguments,\n"+
			"    comma-separated list with -pkgs, or all stdlib with -pkgs std.\n\n"+
			"OPTIONS\n", progName)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES\n"+
			"    %[1]s fmt\n"+
			"        Fetch importers for the fmt package\n\n"+
			"    %[1]s fmt bufio net/http golang.org/x/tools/go/analysis\n"+
			"        Fetch importers for multiple packages\n\n"+
			"    %[1]s -pkgs fmt,bufio,net/http\n"+
			"        Fetch importers using comma-separated packages\n\n"+
			"    %[1]s -pkgs std\n"+
			"        Fetch importers for all standard library packages\n\n"+
			"    %[1]s -workers 20 -pkgs std\n"+
			"        Use 20 concurrent requests when fetching all stdlib packages\n\n"+
			"    %[1]s -pkgs std -sort count\n"+
			"        Fetch all stdlib packages and sort by importer count descending\n", progName)
	}
	flag.Parse()

	// Validate sort flag
	if *sortBy != "name" && *sortBy != "count" {
		return &cmdError{code: 2, msg: fmt.Sprintf("invalid -sort value: %q (must be 'name' or 'count')", *sortBy)}
	}

	args := flag.Args()

	// Validate input: cannot use both -pkgs and positional arguments
	if *pkgsList != "" && len(args) > 0 {
		return &cmdError{code: 2, msg: "-pkgs and positional arguments cannot be used together"}
	}

	// Validate input: must provide at least one
	if *pkgsList == "" && len(args) == 0 {
		return &cmdError{code: 2, msg: "no packages specified; use -h for help"}
	}

	pkgPaths, err := resolvePackages(*pkgsList, args)
	if err != nil {
		return err
	}

	results, err := fetchImporterCounts(context.Background(), pkgPaths, *workers)
	if err != nil {
		return err
	}

	switch *sortBy {
	case "name":
		slices.SortFunc(results, func(a, b pkgImporter) int {
			return cmp.Compare(a.path, b.path)
		})
	case "count":
		slices.SortFunc(results, func(a, b pkgImporter) int {
			// Sort descending by count, then by name for ties
			return cmp.Or(cmp.Compare(b.count, a.count), cmp.Compare(a.path, b.path))
		})
	}

	// Find max width for alignment
	maxWidth := 0
	for _, importer := range results {
		if len(importer.path) > maxWidth {
			maxWidth = len(importer.path)
		}
	}

	// Ensure at least 20 characters for better readability
	if maxWidth < 20 {
		maxWidth = 20
	}

	for _, importer := range results {
		if _, err := fmt.Fprintf(os.Stdout, "%-*s %s\n", maxWidth, importer.path, formatCount(importer.count)); err != nil {
			return err
		}
	}

	return nil
}

// resolvePackages resolves packages from either the -pkgs flag or positional arguments.
// It handles the special case of "std" to load all standard library packages.
// Caller must ensure that exactly one of pkgsList or args is non-empty.
func resolvePackages(pkgsList string, args []string) ([]string, error) {
	const stdKeyword = "std"

	// Handle positional arguments (including "std")
	if len(args) > 0 {
		if len(args) == 1 && strings.TrimSpace(args[0]) == stdKeyword {
			return loadStdPackagePaths()
		}
		return args, nil
	}

	// Handle -pkgs flag (including "std")
	trimmed := strings.TrimSpace(pkgsList)
	if trimmed == stdKeyword {
		return loadStdPackagePaths()
	}
	pkgs := strings.Split(trimmed, ",")
	for i := range pkgs {
		pkgs[i] = strings.TrimSpace(pkgs[i])
	}
	return pkgs, nil
}

// fetchImporterCounts fetches the number of known importers for each package in pkgPaths
// concurrently using the specified number of workers.
// It returns a slice of pkgImporter with package paths and their importer counts.
func fetchImporterCounts(ctx context.Context, pkgPaths []string, workers int) ([]pkgImporter, error) {
	jobs := make(chan string, len(pkgPaths))
	results := make(map[string]int)
	var mu sync.Mutex

	client := &http.Client{}
	// Rate limiter: 1 request per second with burst of 3
	limiter := rate.NewLimiter(rate.Every(time.Second), 3)

	g, gctx := errgroup.WithContext(ctx)
	for range workers {
		g.Go(func() error {
			for path := range jobs {
				// Wait for rate limiter before making request
				if err := limiter.Wait(gctx); err != nil {
					return err
				}

				// Add random jitter (50-200ms) to make pattern less predictable
				jitter := 50*time.Millisecond + rand.N(150*time.Millisecond)
				select {
				case <-time.After(jitter):
				case <-gctx.Done():
					return gctx.Err()
				}

				reqCtx, cancel := context.WithTimeout(gctx, 15*time.Second)
				count, err := fetchImporterCount(reqCtx, client, path)
				cancel()

				if err != nil {
					return fmt.Errorf("fetch %s: %w", path, err)
				}

				mu.Lock()
				results[path] = count
				mu.Unlock()
			}
			return nil
		})
	}

	for _, path := range pkgPaths {
		jobs <- path
	}
	close(jobs)

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Convert results map to slice of pkgImporter
	importers := make([]pkgImporter, 0, len(pkgPaths))
	for _, path := range pkgPaths {
		if count, ok := results[path]; ok {
			importers = append(importers, pkgImporter{path: path, count: count})
		}
	}
	return importers, nil
}

var importerRe = regexp.MustCompile(`Known importers:\s*</strong>\s*([\d,]+)`)

// fetchImporterCount retrieves the number of known importers for a Go package
// from pkg.go.dev by scraping the "importedby" tab. E.g., https://pkg.go.dev/io?tab=importedby.
// It returns the count as an integer, or 0 if the count is not found on the page.
func fetchImporterCount(ctx context.Context, client *http.Client, pkgPath string) (int, error) {
	url := "https://pkg.go.dev/" + pkgPath + "?tab=importedby"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Only read first 100KB since "Known importers" appears early in HTML
	limitedReader := io.LimitReader(resp.Body, 100*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return 0, fmt.Errorf("read body: %w", err)
	}

	m := importerRe.FindSubmatch(body)
	if m == nil {
		return 0, nil
	}

	// Remove commas from the count string before parsing
	countStr := strings.ReplaceAll(string(m[1]), ",", "")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("parse count: %w", err)
	}
	return count, nil
}

// loadStdPackagePaths returns a list of all standard library package paths, excluding internal and vendor packages.
func loadStdPackagePaths() ([]string, error) {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		return nil, fmt.Errorf("load std packages: %w", err)
	}

	var paths []string
	for _, pkg := range pkgs {
		if isInternalOrVendorPackage(pkg.PkgPath) {
			continue
		}
		paths = append(paths, pkg.PkgPath)
	}
	return paths, nil
}

// isInternalOrVendorPackage reports whether the path represents an internal or vendor directory.
func isInternalOrVendorPackage(path string) bool {
	for p := range strings.SplitSeq(path, "/") {
		if p == "internal" || p == "vendor" {
			return true
		}
	}
	return false
}

// formatCount returns a human-friendly string representation of a number with comma separators.
func formatCount(n int) string {
	str := strconv.Itoa(n)
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}
