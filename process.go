package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	".nim":   "Nim",
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

	if filepath.Base(path) == "README_EN.md" {
		pr.processDir(path)
	}

	return nil
}

func (pr *Processor) processDir(path string) {
	dir := filepath.Dir(path)

	id, title, err := pr.parseDir(dir)
	if err != nil {
		log.Printf("Fail to parse id and title: %s", filepath.Base(dir))
		return
	}

	difficulty, tags, description, err := pr.parseReadme(path)
	if err != nil {
		log.Printf("Error reading README %s: %v", path, err)
		return
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dir, err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "Solution.") {
			continue
		}

		fileName := file.Name()
		ext := filepath.Ext(fileName)
		lang, ok := extensionToLanguage[ext]
		if !ok {
			log.Printf("Unknown language for solution file %s/%s: %s", dir, fileName, ext)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			log.Printf("Error reading solution file %s: %v", fileName, err)
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

func (pr *Processor) parseDir(path string) (id int64, title string, err error) {
	base := filepath.Base(path)

	titleRegex := regexp.MustCompile(`^(\d+)\.(.+)$`)
	if matches := titleRegex.FindStringSubmatch(base); matches != nil {
		title = matches[2]
		idStr := matches[1]
		id, err = strconv.ParseInt(idStr, 10, 0)
	} else {
		err = errors.New("title does not match")
	}
	return id, title, err
}

func (pr *Processor) parseReadme(path string) (difficulty string, tags []string, description string, err error) {
	readme, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading README %s: %v", path, err)
		return difficulty, tags, description, err
	}

	content := string(readme)

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
	return difficulty, tags, description, err
}
