# folder-bundler v2.0

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
./bundler collect -compress ./myproject
```

Recreate the structure elsewhere:
```bash
./bundler reconstruct project_collated.md
```

## Core Features

The tool creates comprehensive project documentation including file contents, directory structures, and metadata. It handles text and binary files appropriately, supports syntax highlighting for major programming languages, and manages large projects through automatic file splitting.

### Compression Support (New in v2.0)

folder-bundler now includes advanced compression strategies using hexagonal architecture:

- **Dictionary Compression**: Finds and replaces repeated patterns (up to 89% reduction)
- **Template Compression**: Identifies and parameterizes similar code structures
- **Delta Compression**: Stores files as differences from similar base files
- **Combined Compression**: Layers multiple strategies for maximum compression

When reconstructing projects, it accurately recreates the original structure while preserving file contents, metadata, and timestamps. Compression is automatically detected and handled during reconstruction.

## Configuration Options

Customize collection with these parameters:
```bash
./bundler collect -max-file 5M -exclude-dirs "logs,temp" -compress -compression auto ./myproject
```

Common settings:
- `-max-file`: Set maximum file size (default: 2MB)
- `-exclude-dirs`: Skip specific directories
- `-include-hidden`: Include hidden files
- `-preserve-time`: Keep original timestamps during reconstruction
- `-compress`: Enable compression (default: false)
- `-compression`: Choose strategy: none|auto|dictionary|template|delta|template+delta (default: auto)

The tool automatically excludes common directories like node_modules, dist, and build, as well as binary files (.exe, .dll, etc.) and lock files.

## Compression Examples

```bash
# Auto-select best compression strategy
./bundler collect -compress ./myproject

# Use specific compression strategy
./bundler collect -compress -compression dictionary ./docs
./bundler collect -compress -compression template ./src
./bundler collect -compress -compression delta ./configs

# Use combined compression for maximum reduction
./bundler collect -compress -compression template+delta ./myproject
```

## Use Cases

folder-bundler works well for:
- Project documentation and archiving
- Deployment preparation
- Code review preparation
- Project structure analysis
- Reducing storage size of code archives
- Efficient code distribution
