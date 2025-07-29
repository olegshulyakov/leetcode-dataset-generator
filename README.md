# LeetCode Solutions Dataset Generator

![Go Version](https://img.shields.io/badge/go-1.18%2B-blue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command-line tool to generate Hugging Face datasets from [Doocs LeetCode](https://github.com/doocs/leetcode) solutions repository. Creates structured datasets for fine-tuning large language models with LeetCode problems and solutions.

## Features

- 🚀 Extract LeetCode solutions with metadata
- 📊 Multiple output formats: Parquet (default), CSV, JSON
- 📚 Parses problem descriptions, difficulties, and tags
- 💾 Efficient processing with streaming writes
- 🔍 Automatic language detection from file extensions

## Installation

### Prerequisites

- Go 1.18 or higher
- Git (to clone the repository)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/olegshulyakov/leetcode-dataset-generator.git
cd leetcode-dataset-generator

# Install dependencies
go mod tidy

# Build the tool
go build -o leetcode-dataset
```

### Download Pre-built Binary

Check the [Releases page](https://github.com/olegshulyakov/leetcode-dataset-generator/releases) for pre-built binaries.

## Usage

### Basic Example

```bash
# Clone the Doocs LeetCode repository
git clone https://github.com/doocs/leetcode.git

# Generate dataset in default Parquet format
./leetcode-dataset --repo=leetcode --output=leetcode-solutions
```

### Command Options

```
Usage: leetcode-dataset [options]

Options:
  --repo string      Path to leetcode repository (default ".")
  --convert string   Output format: parquet, csv, or json (default "parquet")
  --output string    Base output filename (default "leetcode-solutions")
  -h, --help         Display help information
```

### Advanced Examples

```bash
# Generate CSV dataset
./leetcode-dataset --repo=leetcode --convert=csv --output=leetcode-csv

# Generate JSON dataset
./leetcode-dataset --repo=leetcode --convert=json --output=leetcode-json

# Specify custom repository path
./leetcode-dataset --repo=/path/to/leetcode --output=my-solutions
```

## Dataset Schema

The generated dataset contains the following columns:

| Column        | Type   | Description                                          |
| ------------- | ------ | ---------------------------------------------------- |
| `id`          | string | Problem ID (e.g., "0001")                            |
| `title`       | string | Problem title (e.g., "two-sum")                      |
| `difficulty`  | string | Problem difficulty ("Easy", "Medium", "Hard")        |
| `description` | string | Problem description in markdown format               |
| `tags`        | string | List of problem tags (e.g., ["Array", "Hash Table"]) |
| `language`    | string | Programming language of solution                     |
| `solution`    | string | Complete solution code                               |

## Supported Languages

The tool automatically detects these programming languages based on file extensions:

- `.c` → C
- `.cpp` → C++
- `.cs` → C#
- `.go` → Go
- `.java` → Java
- `.js` → JavaScript
- `.php` → PHP
- `.py` → Python
- `.rb` → Ruby
- `.rs` → Rust
- `.sh` → Bash
- `.sql` → SQL
- `.ts` → TypeScript
- Other → Uses file extension as language name

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Doocs LeetCode](https://github.com/doocs/leetcode) for the solution repository

---

**Note**: This tool is not affiliated with LeetCode or Doocs. It's built for educational purposes.
