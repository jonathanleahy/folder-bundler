# folder-bundler v3.2

folder-bundler is a Go tool that helps you document and recreate project file structures. It creates detailed documentation of your project files and allows you to rebuild the structure elsewhere, with optional compression to reduce file sizes.

## Quick Start

Install the tool:
```bash
git clone [repository-url]
cd folder-bundler
go build -o bundler
```

Document your project structure:
```bash
./bundler collect ./myproject
```

Document with compression:
```bash
./bundler collect -compress auto ./myproject
```

Recreate the structure elsewhere:
```bash
./bundler reconstruct project_collated_part1.fb
```

## Core Features

The tool creates comprehensive project documentation including file contents, directory structures, and metadata. Output files use the `.fb` extension (folder bundle) to avoid editor encoding issues. 

**Binary File Support (v3.1+)**: Binary files (.ico, .jpg, .mp3, .wav, etc.) are automatically encoded to base64 for storage and decoded during reconstruction, ensuring perfect reproduction of all file types. Large files are handled with proper line wrapping (v3.2).

The tool supports syntax highlighting for major programming languages, manages large projects through automatic file splitting, and calculates SHA-256 hashes for all files to ensure accurate reconstruction.

### Compression Support

folder-bundler now includes advanced compression strategies using hexagonal architecture:

- **Dictionary Compression**: Finds and replaces repeated patterns (up to 89% reduction)
  - v3.0: Now 10-100x faster with intelligent sampling and optimized algorithms
- **Template Compression**: Identifies and parameterizes similar code structures
- **Delta Compression**: Stores files as differences from similar base files
- **Combined Compression**: Layers multiple strategies for maximum compression

When reconstructing projects, it accurately recreates the original structure while preserving file contents, metadata, and timestamps. Compression is automatically detected and handled during reconstruction. All files are verified using SHA-256 hashes to ensure they match the original content exactly.

## Configuration Options

Customize collection with these parameters:
```bash
./bundler collect -max 5M -skip-dirs "logs,temp" -compress auto ./myproject
```

Common settings:
- `-max`: Maximum file size (default: 2M, accepts: 500K, 1M, 2G)
- `-out-max`: Maximum output file size (default: 2M)
- `-skip-dirs`: Skip directories (default: node_modules,.git,...)
- `-skip-files`: Skip files (default: .DS_Store,.env,...)
- `-skip-ext`: Skip extensions (default: .exe,.dll,...)
- `-hidden`: Include hidden files (default: false)
- `-no-gitignore`: Skip .gitignore (default: false)
- `-time`: Preserve timestamps (default: true)
- `-compress`: Compression: none|auto|dictionary|template|delta|template+delta (default: none)

The tool automatically excludes common directories like node_modules, dist, and build, as well as binary files (.exe, .dll, etc.) and lock files.

## Compression Examples

```bash
# No compression (default)
./bundler collect ./myproject

# Auto-select best compression strategy
./bundler collect -compress auto ./myproject

# Use specific compression strategy
./bundler collect -compress dictionary ./docs
./bundler collect -compress template ./src
./bundler collect -compress delta ./configs

# Use combined compression for maximum reduction
./bundler collect -compress template+delta ./myproject

# Flags can be placed anywhere
./bundler collect ./myproject -compress auto
./bundler collect -compress dictionary ./docs -max 5M
```

## Use Cases

folder-bundler works well for:
- Project documentation and archiving
- Deployment preparation
- Code review preparation
- Project structure analysis
- Reducing storage size of code archives
- Efficient code distribution

## Changelog

### v3.2
- **Fixed Large Binary File Support**: Resolved "token too long" error for large files
  - Base64 output now wrapped at 76 characters per line
  - Handles binary files of any size without scanner limitations
  - Maintains backward compatibility with v3.1 files

### v3.1
- **Binary File Support**: All binary files are now included in bundles
  - Binary files (.ico, .jpg, .mp3, .wav, etc.) are automatically base64 encoded
  - Perfect reconstruction of binary files with SHA-256 verification
  - Seamless handling - no configuration needed
  - Works with images, audio, video, executables, and any other binary format

### v3.0
- **Major Performance Improvement**: Dictionary compression is now 10-100x faster for large files
  - Implemented intelligent sampling for files over 50KB
  - Added pattern count limits to prevent excessive memory usage
  - Optimized pattern detection algorithm from O(n²) to ~O(n)
  - Added early termination for unlikely compression candidates
- Performance optimizations maintain compression quality while drastically reducing processing time

### v2.9
- Fixed `-max` flag to properly enforce file size limits
- Fixed UTF-8 handling in template compression to prevent panics
- Improved error messages to show actual file sizes when skipping
