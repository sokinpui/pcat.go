package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"pcat/internal/clipboard"
	"pcat/internal/pcat"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, " ")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func parseArgs(args []string) ([]string, error) {
	var pathArgs []string
	var flagArgs []string

	// Flags that take a value.
	valueFlags := map[string]bool{
		"e":         true,
		"extension": true,
		"not":       true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Arguments not starting with '-' are considered paths.
		if !strings.HasPrefix(arg, "-") {
			pathArgs = append(pathArgs, arg)
			continue
		}

		flagArgs = append(flagArgs, arg)

		// Strip prefixes and get flag name to check if it's a value flag.
		name := strings.TrimLeft(arg, "-")
		if eqIndex := strings.Index(name, "="); eqIndex != -1 {
			name = name[:eqIndex]
		}

		// If it's a value flag and value is not attached with '=',
		// the next argument is the value, provided it's not a flag itself.
		if valueFlags[name] && !strings.Contains(arg, "=") {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flagArgs = append(flagArgs, args[i+1])
				i++ // Consume the value argument.
			}
		}
	}

	if err := flag.CommandLine.Parse(flagArgs); err != nil {
		return nil, err
	}

	return pathArgs, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "pcat: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var extensions, excludePatterns stringSlice
	flag.Var(&extensions, "e", "Filter by file extensions (e.g., 'py', 'js'). Can be repeated.")
	flag.Var(&extensions, "extension", "Alias for -e.")
	flag.Var(&excludePatterns, "not", "Exclude files matching glob patterns. Can be repeated.")

	withLineNumbers := flag.Bool("n", false, "Include line numbers for each file.")
	withLineNumbersAlias := flag.Bool("with-line-numbers", false, "Alias for -n.")
	hidden := flag.Bool("hidden", false, "Include hidden files and directories.")
	listOnly := flag.Bool("l", false, "List the files that would be processed, without printing content.")
	listOnlyAlias := flag.Bool("list", false, "Alias for -l.")
	toClipboard := flag.Bool("c", false, "Copy the output to the clipboard instead of printing to stdout.")
	toClipboardAlias := flag.Bool("clipboard", false, "Alias for -c.")

	flag.Usage = printUsage

	pathArgs, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	paths, err := getPaths(pathArgs)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		flag.Usage()
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
		WithLineNumbers: *withLineNumbers || *withLineNumbersAlias,
		Hidden:          *hidden,
		ListOnly:        *listOnly || *listOnlyAlias,
		ToClipboard:     *toClipboard || *toClipboardAlias,
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
		var paths []string
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
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: %s [OPTIONS] [PATH...]\n", os.Args[0])
	fmt.Fprintln(out, "Concatenate and print files from specified paths (files and directories).")
	fmt.Fprintln(out, "If no paths are provided, paths are read from stdin.")
	fmt.Fprintln(out, "\nOptions:")
	flag.PrintDefaults()
	fmt.Fprintln(out, "\nExamples:")
	fmt.Fprintln(out, "  pcat ./src ./README.md    # Process all files in ./src and the specific file")
	fmt.Fprintln(out, "  pcat ./src -e py -e js     # Process .py and .js files in ./src")
	fmt.Fprintln(out, "  pcat . --hidden           # Process all files in current dir, including hidden ones")
	fmt.Fprintln(out, "  pcat ./src --not '*_test.py' # Exclude test files")
	fmt.Fprintln(out, "  fd . -e py | pcat         # Process python files found by fd from stdin")
}
