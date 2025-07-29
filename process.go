package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var extensionToLanguage = map[string]string{
	".c":     "C",
	".cj":    "Cangjie",
	".cpp":   "C++",
	".cs":    "C#",
	".dart":  "Dart",
	".go":    "Go",
	".java":  "Java",
	".js":    "JavaScript",
	".kt":    "Kotlin",
	".nim":    "Nim",
	".php":   "PHP",
	".py":    "Python",
	".rb":    "Ruby",
	".rs":    "Rust",
	".scala": "Scala",
	".sh":    "Bash",
	".sql":   "SQL",
	".swift": "Swift",
	".ts":    "TypeScript",
}

type Processor struct {
	writer *DataWriter
}

func (pr *Processor) Process() error {
	root := filepath.Join(*repoPath, "solution")
	return filepath.WalkDir(root, pr.walkDir)
}

func (pr *Processor) walkDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		return nil
	}

	dir := filepath.Dir(path)
	if filepath.Base(path) == "README_EN.md" {
		base := filepath.Base(dir)
		parts := strings.SplitN(base, ".", 2)
		if len(parts) != 2 {
			log.Printf("Fail to parse id and title: %s", base)
			return nil
		}

		id, title := parts[0], parts[1]
		pr.processDir(dir, id, title)
	}

	return nil
}

func (pr *Processor) processDir(dir, id, title string) {
	readmePath := filepath.Join(dir, "README_EN.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		log.Printf("Error reading README in %s: %v", dir, err)
		return
	}

	difficulty, tags, description := pr.parseReadme(string(readme))

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dir, err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, "Solution.") {
			ext := filepath.Ext(name)
			lang, ok := extensionToLanguage[ext]
			if !ok {
				log.Printf("Unknown language for solution file %s/%s: %s", dir, name, ext)
				continue
			}

			content, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				log.Printf("Error reading solution file %s: %v", name, err)
				continue
			}

			record := Record{
				ID:          id,
				Title:       title,
				Difficulty:  difficulty,
				Description: description,
				Tags:        tags,
				Language:    lang,
				Solution:    string(content),
			}

			if err = (*pr.writer).WriteRecord(record); err != nil {
				log.Printf("Error writing record: %v", err)
			}
		}
	}
}

func (pr *Processor) parseReadme(content string) (difficulty string, tags []string, description string) {
	lines := strings.Split(content, "\n")
	difficultyRegex := regexp.MustCompile(`!\[(Easy|Medium|Hard)\]`)
	tagRegex := regexp.MustCompile(`!\[([^\]]+)\]`)

	inDescription := false
	descLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "## Description" {
			inDescription = true
			continue
		}

		if inDescription {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			descLines = append(descLines, line)
		} else {
			if matches := difficultyRegex.FindStringSubmatch(line); matches != nil {
				difficulty = matches[1]
			}

			if tagMatches := tagRegex.FindAllStringSubmatch(line, -1); tagMatches != nil {
				for _, match := range tagMatches {
					if match[1] != "Easy" && match[1] != "Medium" && match[1] != "Hard" {
						tags = append(tags, match[1])
					}
				}
			}
		}
	}

	description = strings.TrimSpace(strings.Join(descLines, "\n"))
	return difficulty, tags, description
}
