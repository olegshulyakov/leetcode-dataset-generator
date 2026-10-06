---
license: cc-by-sa-4.0
task_categories:
- text-generation
language:
- en
tags:
- code
pretty_name: Doocs LeetCode Solutions
size_categories:
- 10K<n<100K
configs:
- config_name: default
  data_files:
  - split: train
    path: default/train/*.parquet
- config_name: problems
  data_files:
  - split: train
    path: problems/train/*.parquet
---

# Doocs LeetCode Solutions

LeetCode problems with solutions in many programming languages, built from the [Doocs LeetCode](https://github.com/doocs/leetcode) repository. Each solution comes with the approach name, the reasoning that leads to it, and an explanation with complexity analysis. The dataset is meant for fine-tuning and evaluating code generation models.

The dataset is regenerated weekly from the latest Doocs commit by the [leetcode-dataset-generator](https://github.com/olegshulyakov/leetcode-dataset-generator) tool.

## Structure

The dataset has two configs:

- `default`: one row per solution file (a problem in one language for one approach).
- `problems`: one row per problem; the solution fields below are nested in a `solutions` list.

### Data Fields

| Field           | Type           | Description                                                  | Example                           |
| --------------- | -------------- | ------------------------------------------------------------ | --------------------------------- |
| `id`            | `int64`        | Problem number                                               | `1`                               |
| `title`         | `string`       | Problem title                                                | "Two Sum"                         |
| `slug`          | `string`       | Problem slug                                                 | "two-sum"                         |
| `url`           | `string`       | Problem URL on leetcode.com                                  | "https://leetcode.com/problems/two-sum" |
| `difficulty`    | `string`       | Difficulty level                                             | "Easy", "Medium", "Hard"          |
| `rating`        | `int64`        | Contest rating; null when the problem has none               | `1249`                            |
| `source`        | `string`       | Contest source; may be empty                                 | "Weekly Contest 379 Q1"           |
| `description`   | `string`       | Problem description as HTML                                  | "<p>Given an array of integers…"  |
| `description_zh` | `string`     | Chinese problem description as HTML; empty in this release   | ""                                |
| `tags`          | `list<string>` | Problem tags                                                 | ["Array", "Hash Table"]           |
| `language`      | `string`       | Programming language of the solution                         | "Python", "Java", "C++"           |
| `approach`      | `int64`        | Approach number within the problem                           | `1`, `2`                          |
| `approach_name` | `string`       | Approach name; may be empty                                  | "Hash Table"                      |
| `thinking`      | `string`       | Reasoning that leads to the approach; may be empty           | "Checking every pair is…"         |
| `explanation`   | `string`       | Approach explanation with complexity analysis; may be empty  | "We can use a hash table…"        |
| `solution`      | `string`       | Complete solution code                                       | "class Solution:\n    def…"       |

Languages: Bash, C, C#, C++, Cangjie, Dart, Go, Java, JavaScript, Kotlin, Nim, PHP, Python, Ruby, Rust, Scala, SQL, Swift, TypeScript.

### Data Splits

A single `train` split contains all problems and solutions. The same problem appears in several rows (one per language and approach), so split by `id` to avoid leakage between your train and test sets.

## How to Use

```python
from datasets import load_dataset

dataset = load_dataset("olegshulyakov/doocs-leetcode-solutions", split="train")
# One row per problem with nested solutions:
# problems = load_dataset("olegshulyakov/doocs-leetcode-solutions", "problems", split="train")

sample = dataset[0]
print(f"Problem {sample['id']}: {sample['title']} ({sample['difficulty']})")
print(f"Approach {sample['approach']}: {sample['approach_name']}")
print(sample["thinking"])
print(sample["solution"])
```

## Limitations

- Solutions are community-written and may not be optimal.
- Descriptions are HTML copied from the problem pages.
- Alternative approaches (`approach` > 1) often have only `thinking` and no `explanation`.
- Coverage is limited to problems that have solutions in the Doocs repository.

## License and Attribution

This dataset is derived from [doocs/leetcode](https://github.com/doocs/leetcode) by the Doocs community and is distributed under the same **[CC-BY-SA-4.0](https://creativecommons.org/licenses/by-sa/4.0/)** license. If you share or adapt it, you must give appropriate credit and distribute your contributions under the same license.

Problem statements are the property of [LeetCode](https://leetcode.com). This dataset is not affiliated with or endorsed by LeetCode or Doocs.

## Contact

For questions or issues, please open an issue on [GitHub](https://github.com/olegshulyakov/leetcode-dataset-generator/issues).
