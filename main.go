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
	"github.com/flaticols/resetgen/internal/types"
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

	var structFilter map[string]bool
	if structsFlag != "" {
		structFilter = make(map[string]bool)
		names := strings.Split(structsFlag, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}

			if strings.Contains(name, ".") {
				parts := strings.Split(name, ".")
				if len(parts) != 2 {
					fmt.Fprintf(os.Stderr, "resetgen: invalid format %s (use Package.Struct)\n", name)
					os.Exit(1)
				}
				pkgPath := parts[0]
				structName := parts[1]

				if !isValidGoIdentifier(structName) {
					fmt.Fprintf(os.Stderr, "resetgen: invalid struct name in %s: %s\n", name, structName)
					os.Exit(1)
				}

				if !isValidPackagePath(pkgPath) {
					fmt.Fprintf(os.Stderr, "resetgen: invalid package path in %s: %s\n", name, pkgPath)
					os.Exit(1)
				}

				structFilter[name] = true
			} else {
				if !isValidGoIdentifier(name) {
					fmt.Fprintf(os.Stderr, "resetgen: invalid struct name: %s\n", name)
					os.Exit(1)
				}
				structFilter[name] = true
			}
		}

		if len(structFilter) == 0 {
			fmt.Fprintln(os.Stderr, "resetgen: -structs flag is empty, nothing to process")
			os.Exit(0)
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		if gofile := os.Getenv("GOFILE"); gofile != "" {
			args = []string{gofile}
		} else {
			args = []string{"."}
		}
	}

	if err := run(args, dryRun, structFilter); err != nil {
		fmt.Fprintf(os.Stderr, "resetgen: %v\n", err)
		os.Exit(1)
	}
}

// run processes Go files found by the given patterns and generates Reset() methods for structs
// that match the structFilter (or have reset tags/directives if no filter is provided).
// If dryRun is true, generated code is printed instead of written to files.
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

// findFiles resolves file patterns (e.g., "./...", "./pkg", "file.go") to a list of Go source files.
// Patterns ending with "/..." recursively walk the directory tree. Hidden directories, vendor,
// and testdata directories are skipped. Test files and generated files are excluded.
func findFiles(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
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

		info, err := os.Stat(pattern)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
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
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, ".gen.go") {
		return false
	}
	return true
}

// processFile parses a Go file, applies struct filtering, generates Reset() methods,
// and writes the result to a .gen.go file. Returns true if at least one struct was processed.
// Parsed structs are filtered by structFilter if provided; otherwise all structs with
// reset tags or directives are processed.
func processFile(path string, dryRun bool, structFilter map[string]bool) (bool, error) {
	info, err := parser.ParseFile(path, nil)
	if err != nil {
		return false, err
	}

	if structFilter != nil && len(info.Structs) > 0 {
		var filteredStructs []types.StructInfo
		for _, s := range info.Structs {
			if shouldProcessStruct(s.Name, info.PkgName, structFilter) {
				filteredStructs = append(filteredStructs, s)
			}
		}
		info.Structs = filteredStructs
	}

	if len(info.Structs) == 0 {
		return false, nil
	}

	if structFilter != nil {
		warnUnfoundStructs(info, structFilter)
	}

	code := generator.Generate(info)
	if code == "" {
		return false, nil
	}

	formatted, err := format.Source([]byte(code))
	if err != nil {
		formatted = []byte(code)
	}

	if dryRun {
		fmt.Printf("// Generated from %s\n", path)
		fmt.Println(string(formatted))
		return true, nil
	}

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

func shouldProcessStruct(structName, pkgName string, filter map[string]bool) bool {
	if filter == nil {
		return true
	}

	if filter[structName] {
		return true
	}

	qualifiedName := pkgName + "." + structName
	return filter[qualifiedName]
}

// warnUnfoundStructs emits warnings for structs specified in the filter but not found in the file.
// Only warns for entries relevant to this file's package; qualified names are only warned if
// they match this package, while simple names always trigger warnings if not found.
func warnUnfoundStructs(info *types.FileInfo, structFilter map[string]bool) {
	if len(structFilter) == 0 {
		return
	}

	foundNames := make(map[string]bool)
	for _, s := range info.Structs {
		foundNames[s.Name] = true
		foundNames[info.PkgName+"."+s.Name] = true
	}

	for name := range structFilter {
		if strings.Contains(name, ".") {
			parts := strings.Split(name, ".")
			if parts[0] == info.PkgName && !foundNames[name] {
				fmt.Fprintf(os.Stderr, "resetgen: warning: struct %s not found in %s\n", parts[1], info.Path)
			}
		} else if !foundNames[name] {
			fmt.Fprintf(os.Stderr, "resetgen: warning: struct %s not found in %s\n", name, info.Path)
		}
	}
}

func isValidGoIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}

	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}

	for i := 1; i < len(name); i++ {
		c := name[i]
		if (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') &&
			(c < '0' || c > '9') && c != '_' {
			return false
		}
	}

	return true
}

func isValidPackagePath(path string) bool {
	if len(path) == 0 {
		return false
	}

	if path[0] == '.' {
		return false
	}

	for _, c := range path {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') &&
			c != '.' && c != '/' && c != '_' {
			return false
		}
	}

	return true
}
