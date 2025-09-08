# pcat - find Project/ | xargs cat

A tool to print source code from multiple directories and files in markdown format.

Don't relay on `upload files` feature of LLM providers' web UI. Why AI studio doesn't support `.js` files upload?

```
usage: pcat [-h] [-e EXT [EXT ...]] [--not PATTERN [PATTERN ...]] [-n] [--hidden] [-l] [-c] [PATH ...]

Concatenate and print files from specified paths (files and directories).

positional arguments:
  PATH                  A list of files and/or directories to process. If not provided, paths are read from stdin.

options:
  -h, --help            show this help message and exit
  -e, --extension EXT [EXT ...]
                        Filter by file extensions (e.g., 'py', 'js'). Applies to directories.
  --not PATTERN [PATTERN ...]
                        Exclude files matching glob patterns against the full path (e.g., '*/__pycache__/*').
  -n, --with-line-numbers
                        Include line numbers for each file.
  --hidden              Include hidden files and directories (those starting with a dot).
  -l, --list            List the files that would be processed, without printing content.
  -c, --clipboard       Copy the output to the clipboard instead of printing to stdout.

Examples:
  pcat ./src ./README.md    # Process all files in ./src and the specific file
  pcat ./src -e py js       # Process .py and .js files in ./src
  pcat file1.txt file2.txt  # Concatenate specific files
  pcat . --hidden           # Process all files in current dir, including hidden ones
  pcat ./src -e py -n       # Print python files from ./src with line numbers
  pcat ./src --list         # List files that would be processed in ./src
  pcat ./src --not '*_test.py' '*/gen/*' # Exclude test files and generated code
  pcat ./src -c              # Copy content of files in ./src to clipboard
  fd . -e py | pcat         # Process python files found by fd from stdin

```
