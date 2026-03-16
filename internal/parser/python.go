package parser

import (
	"regexp"
	"strings"

	"codetree/internal/model"
)

type PythonParser struct {
	funcRegex   *regexp.Regexp
	classRegex  *regexp.Regexp
	docstringRe *regexp.Regexp
}

func NewPythonParser() *PythonParser {
	return &PythonParser{
		funcRegex:  regexp.MustCompile(`(?m)^(\s*)def\s+(\w+)\s*\(([^)]*)\)(?:\s*->\s*([^\:]+))?\s*:`),
		classRegex: regexp.MustCompile(`(?m)^(\s*)class\s+(\w+)(?:\s*\(([^)]*)\))?\s*:`),
	}
}

func (p *PythonParser) Language() string {
	return "python"
}

func (p *PythonParser) Extensions() []string {
	return []string{".py", ".pyw", ".pyi"}
}

func (p *PythonParser) LibDirs() []string {
	return []string{
		"venv", ".venv", "env", ".env",
		"__pycache__",
		".tox", ".nox",
		".eggs", "*.egg-info",
		".mypy_cache", ".pytest_cache", ".ruff_cache",
		"site-packages",
		"dist", "build", ".pytype",
		".pants.d", ".pvenv",
	}
}

func (p *PythonParser) Parse(content []byte) ([]model.CodeEntity, error) {
	lines := strings.Split(string(content), "\n")
	return p.parseLines(lines), nil
}

func (p *PythonParser) parseLines(lines []string) []model.CodeEntity {
	var entities []model.CodeEntity
	i := 0

	for i < len(lines) {
		line := lines[i]

		if match := p.classRegex.FindStringSubmatch(line); match != nil {
			indent := len(match[1])
			name := match[2]
			bases := match[3]

			sig := "class " + name
			if bases != "" {
				sig += "(" + bases + ")"
			}

			docstring := ""
			methods := []model.CodeEntity{}

			if i+1 < len(lines) {
				docstring, i = p.extractDocstringFromLine(lines, i+1, indent+4)
			}

			startLine := i + 1
			for i < len(lines)-1 {
				i++
				nextLine := lines[i]
				if strings.TrimSpace(nextLine) == "" {
					continue
				}
				nextIndent := p.getIndent(nextLine)
				if nextIndent <= indent {
					i--
					break
				}

				if methodMatch := p.funcRegex.FindStringSubmatch(nextLine); methodMatch != nil {
					methodIndent := len(methodMatch[1])
					if methodIndent == indent+4 {
						methodName := methodMatch[2]
						params := methodMatch[3]
						returnType := strings.TrimSpace(methodMatch[4])

						methodSig := "def " + methodName + "(" + params + ")"
						if returnType != "" {
							methodSig += " -> " + returnType
						}

						methodDoc := ""
						if i+1 < len(lines) {
							methodDoc, i = p.extractDocstringFromLine(lines, i+1, methodIndent+4)
						}

						methods = append(methods, model.CodeEntity{
							Name:      methodName,
							Type:      model.EntityMethod,
							Signature: methodSig,
							Docstring: methodDoc,
							LineStart: i + 1,
						})
					}
				}
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityClass,
				Signature: sig,
				Docstring: docstring,
				LineStart: startLine,
				Children:  methods,
			})
		} else if match := p.funcRegex.FindStringSubmatch(line); match != nil {
			indent := len(match[1])
			name := match[2]
			params := match[3]
			returnType := strings.TrimSpace(match[4])

			if indent > 0 {
				i++
				continue
			}

			sig := "def " + name + "(" + params + ")"
			if returnType != "" {
				sig += " -> " + returnType
			}

			docstring := ""
			if i+1 < len(lines) {
				docstring, i = p.extractDocstringFromLine(lines, i+1, indent+4)
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityFunction,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		}

		i++
	}

	return entities
}

func (p *PythonParser) getIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

func (p *PythonParser) extractDocstringFromLine(lines []string, startIdx int, expectedIndent int) (string, int) {
	if startIdx >= len(lines) {
		return "", startIdx - 1
	}

	line := lines[startIdx]
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, `"""`) && !strings.HasPrefix(trimmed, `'''`) {
		return "", startIdx - 1
	}

	quote := `"""`
	if strings.HasPrefix(trimmed, `'''`) {
		quote = `'''`
	}

	if strings.Count(trimmed, quote) >= 2 {
		doc := strings.Trim(trimmed, quote)
		return strings.TrimSpace(doc), startIdx
	}

	var docLines []string
	doc := strings.TrimPrefix(trimmed, quote)
	if doc != "" {
		docLines = append(docLines, doc)
	}

	i := startIdx + 1
	for i < len(lines) {
		if strings.Contains(lines[i], quote) {
			lastLine := strings.Split(lines[i], quote)[0]
			if lastLine != "" {
				docLines = append(docLines, lastLine)
			}
			return strings.TrimSpace(strings.Join(docLines, "\n")), i
		}
		docLines = append(docLines, strings.TrimSpace(lines[i]))
		i++
	}

	return strings.TrimSpace(strings.Join(docLines, "\n")), i - 1
}

func init() {
	Register(NewPythonParser())
}
