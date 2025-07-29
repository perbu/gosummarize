# GoSummarize

GoSummarize is a simple command-line utility that scans Go code and generates summaries of exported declarations with their documentation. 
Typically, an LLM might use this to make a summary of Go code without burning through too many tokens.

## Features

- Recursively finds all Go files in a specified directory
- Extracts and displays all exported declarations:
  - Functions with full signatures
  - Methods with receivers and full signatures
  - Types with their complete definitions
  - Constants and variables with their types and values
- Includes documentation comments for each exported item
- LLM-friendly output with clear file boundaries

## Installation

```bash
go install github.com/perbu/gosummarize@latest
```

Or clone the repository and build it manually:

```bash
git clone https://github.com/perbu/gosummarize.git
cd gosummarize
go build .
go install .
```

## Usage

To summarize Go code in a directory:

```bash
gosummarize [options] /path/to/go/project
```

### Options

- `-t` : Ignore test files (files ending with `_test.go`)

The tool will scan the directory recursively and print information about all exported declarations in each Go file, including their full signatures and documentation.

### Example Output

```
<<<FILE_START>>> /path/to/file.go

func NewClient(addr string, options ...Option) (*Client, error)
    NewClient creates a new client with the given address and options.

type Client struct {
    // Fields not shown
}

func (*Client) Connect() error
    Connect establishes a connection to the server.

<<<FILE_END>>> /path/to/file.go
```

## Use Cases

- Quickly understanding the public API of a Go package
- Generating documentation for Go libraries
- Preparing Go code summaries for LLMs to assist in code understanding
- Learning the structure of a new Go codebase

## License

GoSummarize is available under the same BSD-style license as Go. See the [LICENSE](LICENSE) file for details.