# LeetCode Solutions Dataset Generator

![Go Version](https://img.shields.io/badge/go-1.26%2B-blue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command-line tool to generate Hugging Face datasets from [Doocs LeetCode](https://github.com/doocs/leetcode) solutions repository. Creates structured datasets for fine-tuning large language models with LeetCode problems and solutions.

## Features

- 🚀 Extract LeetCode solutions with metadata
- 📊 Multiple output formats: Parquet (default), CSV, JSON Lines
- 📚 Parses problem descriptions, difficulties, tags, and per-approach explanations
- 💾 Efficient processing with streaming writes
- 🔍 Automatic language detection from file extensions

## Installation

### Prerequisites

- Go 1.26 or higher
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
  --layout string    Row layout: solutions or problems (default "solutions")
  --split string     Problems to include: all, train, or test (default "all")
  --test-percent int Percentage of problems assigned to the test split (default 5)
  --min-records int  Fail if fewer records are written (default 0)
  --max-failures int Fail if more problems or solution files fail to parse (default -1, disabled)
  --include-zh       Fill description_zh with the Chinese problem description
  -h, --help         Display help information
```

### Advanced Examples

```bash
# Generate CSV dataset
./leetcode-dataset --repo=leetcode --convert=csv --output=leetcode-csv

# Generate JSON Lines dataset (one JSON object per line, written to a .json file)
./leetcode-dataset --repo=leetcode --convert=json --output=leetcode-json

# Specify custom repository path
./leetcode-dataset --repo=/path/to/leetcode --output=my-solutions
```

## Dataset Schema

With `--layout=solutions` (default), each row is one solution file: a problem in one language for one approach.

With `--layout=problems`, each row is one problem; `language`, `approach`, `approach_name`, `thinking`, `explanation`, and `solution` are nested in a `solutions` list. CSV supports only the solutions layout.

| Column          | Type   | Description                                                             |
| --------------- | ------ | ----------------------------------------------------------------------- |
| `id`            | int    | Problem number (e.g., `1`)                                              |
| `title`         | string | Problem title (e.g., "Two Sum")                                         |
| `slug`          | string | Problem slug (e.g., "two-sum")                                          |
| `url`           | string | Problem URL on leetcode.com                                             |
| `difficulty`    | string | Problem difficulty ("Easy", "Medium", "Hard")                           |
| `rating`        | int    | Contest rating; null when the problem has none                          |
| `source`        | string | Contest source (e.g., "Weekly Contest 379 Q1"); may be empty            |
| `description`   | string | Problem description as HTML                                             |
| `description_zh` | string | Chinese problem description as HTML; empty unless `--include-zh`        |
| `tags`          | list   | Problem tags (e.g., ["Array", "Hash Table"]); `; `-joined string in CSV |
| `language`      | string | Programming language of the solution                                    |
| `approach`      | int    | Approach number (`Solution.py` → 1, `Solution2.py` → 2)                 |
| `approach_name` | string | Approach name (e.g., "Sliding Window"); may be empty                    |
| `thinking`      | string | Reasoning that leads to the approach; may be empty                      |
| `explanation`   | string | Approach explanation with complexity; may be empty                      |
| `solution`      | string | Complete solution code                                                  |

## Supported Languages

The tool detects the programming language from the solution file extension:

| Extension | Language   | Extension | Language   |
| --------- | ---------- | --------- | ---------- |
| `.c`      | C          | `.nim`    | Nim        |
| `.cj`     | Cangjie    | `.php`    | PHP        |
| `.cpp`    | C++        | `.py`     | Python     |
| `.cs`     | C#         | `.rb`     | Ruby       |
| `.dart`   | Dart       | `.rs`     | Rust       |
| `.go`     | Go         | `.scala`  | Scala      |
| `.java`   | Java       | `.sh`     | Bash       |
| `.js`     | JavaScript | `.sql`    | SQL        |
| `.kt`     | Kotlin     | `.swift`  | Swift      |
| `.ts`     | TypeScript |           |            |

Files with other extensions are skipped and counted as failures (see `--max-failures`).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Doocs LeetCode](https://github.com/doocs/leetcode) for the solution repository

---

**Note**: This tool is not affiliated with LeetCode or Doocs. It's built for educational purposes.
