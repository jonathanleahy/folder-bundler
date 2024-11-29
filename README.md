# Unified File Structure Manager

A robust Go application for collecting and reconstructing file structures. This tool enables you to create detailed documentation of your project's file structure and reconstruct it elsewhere, making it valuable for project archiving, documentation, and deployment.

## Features

The File Structure Manager provides two primary operations:

Collection creates a detailed Markdown summary of your project structure, including:
- File contents with syntax highlighting for supported languages
- Directory structure preservation
- File metadata (size, modification times)
- Intelligent handling of binary and text files
- Support for large projects through automatic file splitting
- Configurable file exclusions

Reconstruction capabilities include:
- Accurate recreation of directory structures
- Preservation of file contents and metadata
- Support for multi-part file reconstruction
- Optional timestamp preservation
- Detailed progress reporting

## Installation

To use this tool, ensure you have Go installed on your system (version 1.16 or later recommended). Then follow these steps:

1. Clone the repository:
```bash
git clone [repository-url]
cd file-structure-manager
```

2. Install the required dependency:
```bash
go get github.com/sabhiram/go-gitignore
```

3. Build the application:
```bash
go build -o fsmanager
```

## Usage

The application supports two main commands: `collect` and `reconstruct`.

### Collecting File Structure

Basic usage:
```bash
./fsmanager collect [directory_path]
```

With custom parameters:
```bash
./fsmanager collect -max-file 5242880 -max-output 10485760 -exclude-dirs "logs,temp" ./myproject
```

Available collection parameters:
- `-max-file`: Maximum size of individual files to process (bytes, default: 1MB)
- `-max-output`: Maximum size of output files (bytes, default: 2MB)
- `-exclude-dirs`: Comma-separated list of directories to exclude
- `-exclude-files`: Comma-separated list of files to exclude
- `-exclude-exts`: Comma-separated list of file extensions to exclude
- `-include-hidden`: Include hidden files and directories (default: false)
- `-skip-gitignore`: Skip .gitignore processing (default: false)

### Reconstructing File Structure

Basic usage:
```bash
./fsmanager reconstruct <input_file>
```

With custom parameters:
```bash
./fsmanager reconstruct -preserve-time=false project_collated.md
```

Available reconstruction parameters:
- `-preserve-time`: Preserve original timestamps (default: true)

## Default Exclusions

The tool automatically excludes certain files and directories by default:

Directories:
- node_modules
- vendor
- venv
- dist
- build

File Extensions:
- .exe, .dll, .so, .dylib (Binary executables)
- .bin, .pkl, .pyc (Binary data files)
- .bak (Backup files)

Files:
- package-lock.json
- yarn.lock

## Supported Programming Languages

The tool provides syntax highlighting for numerous programming languages in the generated documentation, including:
- Go, Python, JavaScript, TypeScript
- Java, Kotlin, Scala
- C++, C, Rust
- PHP, Ruby, Perl
- HTML, CSS, XML, JSON, YAML
- SQL, R, Swift
- And many more

## Technical Details

The tool implements several important features for reliable file handling:

1. Text File Detection
  - UTF-8 validation
  - Null byte detection
  - Empty file handling

2. Size Management
  - Automatic file splitting for large projects
  - Configurable size limits
  - Skip handling for oversized files

3. Error Handling
  - Comprehensive error reporting
  - Backup creation for existing files
  - Directory permission management

## Common Use Cases

The File Structure Manager is particularly useful for:
- Project documentation and archiving
- Deployment preparation and verification
- Project structure analysis
- Code review preparation
- Project templating

## Best Practices

1. When collecting files:
  - Start with default exclusions and adjust as needed
  - Use appropriate size limits for your project
  - Consider enabling hidden file inclusion for complete documentation

2. When reconstructing:
  - Ensure sufficient disk space at the target location
  - Verify all parts are available for multi-part collections
  - Consider timestamp preservation requirements

## Contributing

We welcome contributions to improve the File Structure Manager. Please ensure you:
1. Follow Go coding standards
2. Add tests for new features
3. Update documentation as needed
4. Submit detailed pull requests

## License

[Add your license information here]