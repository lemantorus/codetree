package parser

import (
	"regexp"
	"strings"

	"codetree/internal/model"
)

type JavaScriptParser struct {
	funcRegex   *regexp.Regexp
	classRegex  *regexp.Regexp
	methodRegex *regexp.Regexp
}

func NewJavaScriptParser() *JavaScriptParser {
	return &JavaScriptParser{
		funcRegex:   regexp.MustCompile(`(?m)(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)`),
		classRegex:  regexp.MustCompile(`(?m)(?:export\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?`),
		methodRegex: regexp.MustCompile(`(?m)^\s*(?:async\s+)?(\w+)\s*\(([^)]*)\)\s*\{`),
	}
}

func (p *JavaScriptParser) Language() string {
	return "javascript"
}

func (p *JavaScriptParser) Extensions() []string {
	return []string{".js", ".mjs", ".cjs"}
}

func (p *JavaScriptParser) LibDirs() []string {
	return []string{
		"node_modules",
		".npm", ".yarn", ".pnpm-store",
		"bower_components", "jspm_packages",
		".cache", ".parcel-cache",
		".nuxt", ".next",
	}
}

func (p *JavaScriptParser) Parse(content []byte) ([]model.CodeEntity, error) {
	lines := strings.Split(string(content), "\n")
	return p.parseLines(lines), nil
}

func (p *JavaScriptParser) parseLines(lines []string) []model.CodeEntity {
	var entities []model.CodeEntity
	classIndent := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if match := p.classRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			sig := "class " + name
			if match[2] != "" {
				sig += " extends " + match[2]
			}
			classIndent = p.getIndent(line)

			docstring := ""
			if i > 0 {
				docstring = p.extractJSDoc(lines, i-1)
			}

			var methods []model.CodeEntity
			for j := i + 1; j < len(lines); j++ {
				nextLine := lines[j]
				if strings.TrimSpace(nextLine) == "" {
					continue
				}
				nextIndent := p.getIndent(nextLine)
				if nextIndent <= classIndent && strings.TrimSpace(nextLine) != "" && !strings.HasPrefix(strings.TrimSpace(nextLine), "//") {
					break
				}

				if methodMatch := p.methodRegex.FindStringSubmatch(nextLine); methodMatch != nil {
					methodName := methodMatch[1]
					if methodName == "constructor" || p.isValidMethodName(methodName) {
						params := methodMatch[2]
						methodSig := methodName + "(" + params + ")"

						methodDoc := p.extractJSDoc(lines, j-1)

						methods = append(methods, model.CodeEntity{
							Name:      methodName,
							Type:      model.EntityMethod,
							Signature: methodSig,
							Docstring: methodDoc,
							LineStart: j + 1,
						})
					}
				}
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityClass,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
				Children:  methods,
			})
		} else if match := p.funcRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			params := match[2]
			sig := "function " + name + "(" + params + ")"

			docstring := ""
			if i > 0 {
				docstring = p.extractJSDoc(lines, i-1)
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityFunction,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		}
	}

	return entities
}

func (p *JavaScriptParser) getIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 2
		} else {
			break
		}
	}
	return count
}

func (p *JavaScriptParser) isValidMethodName(name string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true,
		"return": true, "const": true, "let": true, "var": true,
		"function": true, "class": true, "export": true, "import": true,
	}
	return !keywords[name]
}

func (p *JavaScriptParser) extractJSDoc(lines []string, lineIdx int) string {
	if lineIdx < 0 {
		return ""
	}

	var docLines []string
	foundEnd := false

	for i := lineIdx; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "/**") && strings.HasSuffix(trimmed, "*/") {
			doc := strings.TrimPrefix(trimmed, "/**")
			doc = strings.TrimSuffix(doc, "*/")
			doc = strings.TrimSpace(doc)
			if doc != "" {
				docLines = append([]string{doc}, docLines...)
			}
			break
		}

		if strings.HasPrefix(trimmed, "/**") {
			doc := strings.TrimPrefix(trimmed, "/**")
			doc = strings.TrimSpace(doc)
			if doc != "" {
				docLines = append([]string{doc}, docLines...)
			}
			break
		}

		if strings.HasSuffix(trimmed, "*/") {
			foundEnd = true
			doc := strings.TrimSuffix(trimmed, "*/")
			doc = strings.TrimPrefix(doc, "*")
			doc = strings.TrimSpace(doc)
			if doc != "" {
				docLines = append([]string{doc}, docLines...)
			}
			continue
		}

		if foundEnd || strings.HasPrefix(trimmed, "*") {
			doc := strings.TrimPrefix(trimmed, "*")
			doc = strings.TrimSpace(doc)
			if doc != "" {
				docLines = append([]string{doc}, docLines...)
			}
			continue
		}

		if !strings.HasPrefix(trimmed, "//") && trimmed != "" {
			break
		}
	}

	return strings.Join(docLines, "\n")
}

func init() {
	Register(NewJavaScriptParser())
}
