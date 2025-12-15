package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/flaticols/resetgen/internal/generator"
	"github.com/flaticols/resetgen/internal/parser"
)

func main() {
	var (
		showVersion bool
		dryRun      bool
		structsFlag string
	)

	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&dryRun, "dry-run", false, "print generated code instead of writing files")
	flag.StringVar(&structsFlag, "structs", "", "comma-separated list of struct names to process (e.g., User,Order,Config)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: resetgen [flags] [patterns...]\n\n")
		fmt.Fprintf(os.Stderr, "Generate Reset() methods for structs with reset tags.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  //go:generate resetgen     use in source file (processes that file)\n")
		fmt.Fprintf(os.Stderr, "  resetgen ./...             all packages in current directory tree\n")
		fmt.Fprintf(os.Stderr, "  resetgen ./pkg             specific package\n")
		fmt.Fprintf(os.Stderr, "  resetgen file.go           specific file\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		printVersion()
		return
	}

	// Parse -structs flag into a map for efficient lookup
	var structFilter map[string]bool
	if structsFlag != "" {
		structFilter = make(map[string]bool)
		names := strings.Split(structsFlag, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			// Validate that it's a valid Go identifier
			if !isValidGoIdentifier(name) {
				fmt.Fprintf(os.Stderr, "resetgen: invalid struct name: %s\n", name)
				os.Exit(1)
			}
			structFilter[name] = true
		}

		// Empty list after trimming means process nothing
		if len(structFilter) == 0 {
			fmt.Fprintln(os.Stderr, "resetgen: -structs flag is empty, nothing to process")
			os.Exit(0)
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		// Check for go generate environment
		if gofile := os.Getenv("GOFILE"); gofile != "" {
			// Running via go generate - process current file
			args = []string{gofile}
		} else {
			// Default: process current directory
			args = []string{"."}
		}
	}

	if err := run(args, dryRun, structFilter); err != nil {
		fmt.Fprintf(os.Stderr, "resetgen: %v\n", err)
		os.Exit(1)
	}
}

func run(patterns []string, dryRun bool, structFilter map[string]bool) error {
	files, err := findFiles(patterns)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no Go files found")
	}

	processed := 0
	for _, file := range files {
		ok, err := processFile(file, dryRun, structFilter)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if ok {
			processed++
		}
	}

	if processed == 0 {
		fmt.Fprintln(os.Stderr, "resetgen: no structs with reset tags found")
	}

	return nil
}

func findFiles(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Handle ./... pattern
		if strings.HasSuffix(pattern, "/...") {
			dir := strings.TrimSuffix(pattern, "/...")
			if dir == "." || dir == "" {
				dir = "."
			}
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					// Skip hidden directories and testdata
					name := info.Name()
					if strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor" {
						return filepath.SkipDir
					}
					return nil
				}
				if isGoSourceFile(path) && !seen[path] {
					files = append(files, path)
					seen[path] = true
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}

		// Check if it's a directory
		info, err := os.Stat(pattern)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			// Process all Go files in directory
			entries, err := os.ReadDir(pattern)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				path := filepath.Join(pattern, entry.Name())
				if isGoSourceFile(path) && !seen[path] {
					files = append(files, path)
					seen[path] = true
				}
			}
		} else if isGoSourceFile(pattern) && !seen[pattern] {
			// Single file
			files = append(files, pattern)
			seen[pattern] = true
		}
	}

	return files, nil
}

func isGoSourceFile(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return false
	}
	// Skip test files and generated files
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_test.go") {
		return false
	}
	if strings.HasSuffix(base, ".gen.go") {
		return false
	}
	return true
}

func processFile(path string, dryRun bool, structFilter map[string]bool) (bool, error) {
	info, err := parser.ParseFile(path, structFilter)
	if err != nil {
		return false, err
	}

	if len(info.Structs) == 0 {
		return false, nil
	}

	// Warn about structs that were requested but not found
	if structFilter != nil && len(info.Structs) < len(structFilter) {
		foundNames := make(map[string]bool)
		for _, s := range info.Structs {
			foundNames[s.Name] = true
		}
		for name := range structFilter {
			if !foundNames[name] {
				fmt.Fprintf(os.Stderr, "resetgen: warning: struct %s not found in %s\n", name, path)
			}
		}
	}

	code := generator.Generate(info)
	if code == "" {
		return false, nil
	}

	// Format the generated code
	formatted, err := format.Source([]byte(code))
	if err != nil {
		// If formatting fails, write unformatted code (useful for debugging)
		formatted = []byte(code)
	}

	if dryRun {
		fmt.Printf("// Generated from %s\n", path)
		fmt.Println(string(formatted))
		return true, nil
	}

	// Write to .gen.go file
	outPath := outputPath(path)
	if err := os.WriteFile(outPath, formatted, 0o644); err != nil { //nolint:gosec // generated code should be world-readable
		return false, err
	}

	fmt.Printf("resetgen: wrote %s\n", outPath)
	return true, nil
}

func outputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + ".gen.go"
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, v := range info.Settings {
			if v.Key == "vcs.revision" {
				fmt.Println("resetgen", v.Value)
				return
			}
		}
		fmt.Println("resetgen", info.Main.Version)
		return
	}
	fmt.Println("resetgen", "dev")
}

// isValidGoIdentifier checks if a name is a valid exported Go identifier.
func isValidGoIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Must start with uppercase letter (we only allow exported structs)
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}

	// Rest must be letters, digits, or underscore
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}
