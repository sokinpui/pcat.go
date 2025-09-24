package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/sokinpui/pcat.go/internal/clipboard"
	"github.com/sokinpui/pcat.go/internal/pcat"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "pcat: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var extensions, excludePatterns []string
	var withLineNumbers, hidden, listOnly, toClipboard, noHeader bool

	pflag.StringSliceVarP(&extensions, "extension", "e", nil, "Filter by file extensions (e.g., 'py', 'js'). Can be repeated.")
	pflag.StringSliceVar(&excludePatterns, "not", nil, "Exclude files matching glob patterns. Can be repeated.")
	pflag.BoolVarP(&withLineNumbers, "with-line-numbers", "n", false, "Include line numbers for each file.")
	pflag.BoolVar(&hidden, "hidden", false, "Include hidden files and directories.")
	pflag.BoolVarP(&listOnly, "list", "l", false, "List the files that would be processed, without printing content.")
	pflag.BoolVarP(&toClipboard, "clipboard", "c", false, "Copy the output to the clipboard instead of printing to stdout.")
	pflag.BoolVarP(&noHeader, "no-header", "N", false, "Do not print the header and footer.")

	pflag.Usage = printUsage
	pflag.Parse()

	paths, err := getPaths(pflag.Args())
	if err != nil {
		return err
	}

	if paths == nil {
		pflag.Usage()
		return fmt.Errorf("no paths provided")
	}

	var directories, specificFiles []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("invalid path '%s': %w", path, err)
		}
		if info.IsDir() {
			directories = append(directories, path)
		} else {
			specificFiles = append(specificFiles, path)
		}
	}

	if len(directories) > 0 && len(extensions) == 0 {
		extensions = []string{"any"}
	}

	config := pcat.Config{
		Directories:     directories,
		Extensions:      extensions,
		SpecificFiles:   specificFiles,
		ExcludePatterns: excludePatterns,
		WithLineNumbers: withLineNumbers,
		Hidden:          hidden,
		ListOnly:        listOnly,
		ToClipboard:     toClipboard,
		NoHeader:        noHeader,
	}

	app := pcat.New(config)
	output, err := app.Run()
	if err != nil {
		return err
	}

	if config.ToClipboard {
		if err := clipboard.Write(output); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "pcat: Copied to clipboard.")
	} else if output != "" {
		if _, err := fmt.Fprint(os.Stdout, output); err != nil && !isBrokenPipe(err) {
			return err
		}
	}

	return nil
}

func getPaths(args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		paths := make([]string, 0)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				paths = append(paths, line)
			}
		}
		return paths, scanner.Err()
	}

	return nil, nil
}

func isBrokenPipe(err error) bool {
	// This check is OS-dependent but covers common cases for Linux/macOS.
	return strings.Contains(err.Error(), "broken pipe")
}

func printUsage() {
	out := os.Stderr
	fmt.Fprintf(out, "Usage: %s [OPTIONS] [PATH...]\n", os.Args[0])
	fmt.Fprintln(out, "Concatenate and print files from specified paths (files and directories).")
	fmt.Fprintln(out, "If no paths are provided, paths are read from stdin.")
	fmt.Fprintln(out, "\nOptions:")
	pflag.PrintDefaults()
	fmt.Fprintln(out, "\nExamples:")
	fmt.Fprintln(out, "  pcat ./src ./README.md    # Process all files in ./src and the specific file")
	fmt.Fprintln(out, "  pcat ./src -e py -e js     # Process .py and .js files in ./src")
	fmt.Fprintln(out, "  pcat . --hidden           # Process all files in current dir, including hidden ones")
	fmt.Fprintln(out, "  pcat ./src --not '*_test.py' # Exclude test files")
	fmt.Fprintln(out, "  fd . -e py | pcat         # Process python files found by fd from stdin")
}
