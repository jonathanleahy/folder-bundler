# folder-bundler

folder-bundler is a Go tool that helps you document and recreate project file structures. It creates detailed documentation of your project files and allows you to rebuild the structure elsewhere.

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

Recreate the structure elsewhere:
```bash
./bundler reconstruct project_collated.md
```

## Core Features

The tool creates comprehensive project documentation including file contents, directory structures, and metadata. It handles text and binary files appropriately, supports syntax highlighting for major programming languages, and manages large projects through automatic file splitting.

When reconstructing projects, it accurately recreates the original structure while preserving file contents, metadata, and timestamps.

## Configuration Options

Customize collection with these parameters:
```bash
./bundler collect -max-file 5M -exclude-dirs "logs,temp" ./myproject
```

Common settings:
- `-max-file`: Set maximum file size (default: 1MB)
- `-exclude-dirs`: Skip specific directories
- `-include-hidden`: Include hidden files
- `-preserve-time`: Keep original timestamps during reconstruction

The tool automatically excludes common directories like node_modules, dist, and build, as well as binary files (.exe, .dll, etc.) and lock files.

## Use Cases

folder-bundler works well for:
- Project documentation and archiving
- Deployment preparation
- Code review preparation
- Project structure analysis
