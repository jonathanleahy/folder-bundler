# RecursiveFileCollator

RecursiveFileCollator is a Go program that recursively processes a directory structure and creates a single Markdown file containing the contents of all files, along with their metadata. This tool is particularly useful for creating a comprehensive overview of a project's file structure and contents, which can be easily shared or analyzed.

## Features

- Recursively traverses directory structures
- Ignores files and directories specified in .gitignore
- Excludes common build directories and binary files
- Respects file size limits to avoid processing large files
- Automatically backs up previous output files
- Generates a single Markdown file with:
    - File paths
    - File sizes
    - Last modified dates
    - File contents (with syntax highlighting for supported file types)
- Identifies and skips non-text files
- Handles various input scenarios for the root directory

## Project Structure

```
RecursiveFileCollator/
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

1. Ensure you have Go installed on your system. If not, you can download it from [golang.org](https://golang.org/).

2. Clone this repository:
   ```
   git clone https://github.com/yourusername/RecursiveFileCollator.git
   cd RecursiveFileCollator
   ```

3. Install the required dependency:
   ```
   go get github.com/sabhiram/go-gitignore
   ```

## Usage

1. From the project root directory, run the program with the following command:
   ```
   go run ./src [path_to_directory]
   ```
   Replace `[path_to_directory]` with the path to the directory you want to process. If no path is provided, the current directory will be used.

2. The program will generate a file named `<directory_name>_collated.md` in the current working directory.

3. If a file with the same name already exists, it will be renamed to `<directory_name>_collated.md.bak` before the new file is created.

## Example

If you run:
```
go run ./src /path/to/my/project
```

The program will create a file named `project_collated.md` containing the processed contents of the `/path/to/my/project` directory. If `project_collated.md` already exists, it will be renamed to `project_collated.md.bak`.

## Configuration

You can modify the following variables in the `src/main.go` file to customize the behavior:

- `maxFileSize`: Maximum file size to process (default is 1MB)
- `excludedDirs`: Map of directory names to exclude
- `excludedExtensions`: Map of file extensions to exclude
- `excludedFiles`: Map of specific filenames to exclude

## Running Tests

To run the tests for this program:

1. Ensure you're in the project root directory.
2. Run the following command:
   ```
   go test ./src -v
   ```

## Building the Program

To build an executable:

1. From the project root directory, run:
   ```
   go build -o recursive-file-collate .\src\
   ```

2. This will create an executable named `recursive-file-collate` (or `recursive-file-collate.exe` on Windows) in the project root.

3. You can run the executable with:
   ```
   ./recursive-file-collate [path_to_directory]
   ```

## New Features

- **Automatic Backup**: The program now automatically renames an existing output file to `.bak` before creating a new one.
- **Flexible Input**: You can now run the program without specifying a directory, and it will use the current directory.
- **Non-text File Handling**: The program now identifies and skips non-text files, preventing errors when processing binary files.

## Contributing

Contributions to improve RecursiveFileCollator are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).