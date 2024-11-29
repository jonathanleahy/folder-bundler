# FolderBundler

FolderBundler is a Go program that recursively processes a directory structure and creates Markdown files containing the contents of all files, along with their metadata. This tool is particularly useful for creating a comprehensive overview of a project's file structure and contents, which can be easily shared or analyzed.

## Features

FolderBundler offers comprehensive file processing capabilities with intelligent size management:

- Recursively traverses directory structures
- Automatically splits output into multiple files when exceeding 2MB
- Ignores hidden files and directories (starting with '.')
- Respects patterns specified in .gitignore
- Excludes common build directories and binary files
- Enforces a 1MB limit on individual processed files
- Automatically backs up previous output files
- Generates Markdown files containing:
    - File paths
    - File sizes
    - Last modified dates
    - File contents (with syntax highlighting for supported file types)
- Identifies and skips non-text files
- Handles various input scenarios for the root directory

## Project Structure

```
FolderBundler/
│
├── src/
│   ├── main.go
│   └── main_test.go
│
├── README.md
├── .gitignore
├── go.mod
└── go.sum
```

## Installation

1. Ensure you have Go installed on your system. If not, you can download it from golang.org.
2. Clone this repository:
```bash
git clone https://github.com/yourusername/FolderBundler.git
cd FolderBundler
```

3. Install the required dependency:
```bash
go get github.com/sabhiram/go-gitignore
```

## Usage

1. From the project root directory, run the program with the following command:
```bash
go run ./src [path_to_directory]
```

Replace `[path_to_directory]` with the path to the directory you want to process. If no path is provided, the current directory will be used.

2. The program will generate files named `<directory_name>_collated_part1.md`, `<directory_name>_collated_part2.md`, etc., in the current working directory. New parts are created automatically when a file reaches 2MB in size.

3. If files with the same names already exist, they will be renamed with a `.bak` extension before the new files are created.

## Example

If you run:
```bash
go run ./src /path/to/my/project
```

The program will create files named `project_collated_part1.md`, `project_collated_part2.md`, etc., containing the processed contents of the `/path/to/my/project` directory. If these files already exist, they will be backed up with a `.bak` extension.

## Configuration

You can modify the following variables in the `src/main.go` file to customize the behavior:
- `maxFileSize`: Maximum size of individual files to process (default is 1MB)
- `maxOutputSize`: Maximum size of output files before splitting (default is 2MB)
- `excludedDirs`: Map of directory names to exclude
- `excludedExtensions`: Map of file extensions to exclude
- `excludedFiles`: Map of specific filenames to exclude

## Running Tests

To run the tests for this program:
1. Ensure you're in the project root directory.
2. Run the following command:
```bash
go test ./src -v
```

## Building the Program

To build an executable:
1. From the project root directory, run:
```bash
go build -o folder-bundler ./src/
```

2. This will create an executable named `folder-bundler` (or `folder-bundler.exe` on Windows) in the project root.

3. You can run the executable with:
```bash
./folder-bundler [path_to_directory]
```

## Key Features and Improvements

- **Intelligent File Splitting**: Automatically creates new files when reaching the 2MB size limit
- **Hidden File Handling**: Automatically skips all hidden files and directories (starting with '.')
- **Consistent Formatting**: Maintains proper Markdown formatting across all output files
- **Progress Tracking**: Provides clear console output showing processing status
- **Automatic Backup**: Preserves existing files by creating backups before overwriting
- **Flexible Input**: Supports running without a specified directory
- **Non-text File Handling**: Identifies and skips binary and non-text files

## Contributing

Contributions to improve FolderBundler are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the MIT License.