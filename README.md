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
  --description-format string
                     Description format: html or markdown (default "html")
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
| `description`   | string | Problem description as HTML, or Markdown with `--description-format=markdown` |
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

## Preparing Data for LoRA/QLoRA Fine-Tuning

The published dataset, [olegshulyakov/doocs-leetcode-solutions](https://huggingface.co/datasets/olegshulyakov/doocs-leetcode-solutions), is raw data: one row per solution, with Markdown descriptions and `train`/`test` splits by problem. Turn it into chat examples for your model before fine-tuning:

- Use the prompt-completion format, so the loss covers only the answer and not the problem statement.
- Put the reasoning first: `thinking`, then `explanation`, then the code in a fenced block.
- Keep only the languages you target, and limit approaches per problem so popular problems do not dominate.
- Drop examples longer than your context budget instead of truncating them; truncation cuts off the code.
- Vary the prompt wording, and pick the template by problem `id` so runs are reproducible.

```bash
pip install datasets transformers jinja2
```

````python
from datasets import load_dataset
from transformers import AutoTokenizer

DATASET = "olegshulyakov/doocs-leetcode-solutions"
MODEL = "Qwen/Qwen2.5-Coder-7B-Instruct"  # tokenizer of the model you fine-tune
LANGUAGES = {"Python": "python", "C++": "cpp", "Java": "java"}  # language -> code fence tag
MAX_APPROACHES = 2  # keep at most this many approaches per problem and language
MAX_TOKENS = 4096  # drop longer examples instead of truncating the code
INCLUDE_THINKING = True

SYSTEM = "You are an expert competitive programmer. Reason about the problem first, then write a correct, efficient solution."
PROMPTS = [
    "{description}\n\nSolve this problem in {language}.",
    "Write a {language} solution for the following problem:\n\n{description}",
    "{description}\n\nExplain your approach and implement it in {language}.",
]

tokenizer = AutoTokenizer.from_pretrained(MODEL)


def to_sft(row):
    parts = [row["thinking"]] if INCLUDE_THINKING and row["thinking"] else []
    if row["explanation"]:
        parts.append(row["explanation"])
    parts.append(f"```{LANGUAGES[row['language']]}\n{row['solution'].strip()}\n```")

    # Pick the prompt template by problem id so the choice is stable between runs.
    prompt = PROMPTS[row["id"] % len(PROMPTS)].format(description=row["description"], language=row["language"])
    return {
        "prompt": [{"role": "system", "content": SYSTEM}, {"role": "user", "content": prompt}],
        "completion": [{"role": "assistant", "content": "\n\n".join(parts)}],
    }


def fits(example):
    text = tokenizer.apply_chat_template(example["prompt"] + example["completion"], tokenize=False)
    return len(tokenizer(text)["input_ids"]) <= MAX_TOKENS


dataset = load_dataset(DATASET)  # train and test splits, split by problem
dataset = dataset.filter(lambda r: r["language"] in LANGUAGES and r["approach"] <= MAX_APPROACHES)
dataset = dataset.map(to_sft, remove_columns=dataset["train"].column_names)
dataset = dataset.filter(fits)

print(dataset)
dataset.save_to_disk("leetcode-sft")  # or pass dataset["train"] / dataset["test"] to TRL SFTTrainer
````

The result plugs into TRL's `SFTTrainer` with a `peft_config`. For prompt-completion datasets, the loss is computed on the completion only by default. The dataset is CC-BY-SA-4.0, so credit Doocs when you publish a model or adapter trained on it.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Doocs LeetCode](https://github.com/doocs/leetcode) for the solution repository

---

**Note**: This tool is not affiliated with LeetCode or Doocs. It's built for educational purposes.
