package parser

import (
	"regexp"
	"strings"

	"github.com/lemantorus/codetree/internal/model"
)

type TypeScriptParser struct {
	funcRegex      *regexp.Regexp
	arrowFuncRegex *regexp.Regexp
	classRegex     *regexp.Regexp
	interfaceRegex *regexp.Regexp
	typeRegex      *regexp.Regexp
	enumRegex      *regexp.Regexp
	methodRegex    *regexp.Regexp
}

func NewTypeScriptParser() *TypeScriptParser {
	return &TypeScriptParser{
		funcRegex:      regexp.MustCompile(`(?m)^(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*(?:<[^>]*>)?\s*\(([^)]*)\)(?:\s*:\s*([^{]+))?\s*\{`),
		arrowFuncRegex: regexp.MustCompile(`(?m)^(?:export\s+)?(?:const|let|var)\s+(\w+)\s*(?::\s*[^=]+)?\s*=\s*(?:async\s+)?(?:\([^)]*\)|\w+)\s*(?::\s*[^=]+)?\s*=>`),
		classRegex:     regexp.MustCompile(`(?m)^(?:export\s+)?(?:abstract\s+)?class\s+(\w+)(?:\s*<[^>]*>)?(?:\s+(?:extends|implements)\s+([^{]+))?`),
		interfaceRegex: regexp.MustCompile(`(?m)^(?:export\s+)?interface\s+(\w+)(?:\s*<[^>]*>)?(?:\s+extends\s+([^{]+))?`),
		typeRegex:      regexp.MustCompile(`(?m)^(?:export\s+)?type\s+(\w+)(?:\s*<[^>]*>)?\s*=`),
		enumRegex:      regexp.MustCompile(`(?m)^(?:export\s+)?(?:const\s+)?enum\s+(\w+)`),
		methodRegex:    regexp.MustCompile(`(?m)^\s*(?:abstract\s+)?(?:private\s+|public\s+|protected\s+)?(?:readonly\s+)?(?:async\s+)?(\w+)\s*(?:<[^>]*>)?\s*\(([^)]*)\)(?:\s*:\s*([^{;\n]+))?`),
	}
}

func (p *TypeScriptParser) Language() string {
	return "typescript"
}

func (p *TypeScriptParser) Extensions() []string {
	return []string{".ts", ".tsx", ".mts", ".cts"}
}

func (p *TypeScriptParser) LibDirs() []string {
	return []string{
		"node_modules",
		".npm", ".yarn", ".pnpm-store",
		".cache", ".parcel-cache",
		".nuxt", ".next",
		"dist", "build",
	}
}

func (p *TypeScriptParser) Parse(content []byte) ([]model.CodeEntity, error) {
	lines := strings.Split(string(content), "\n")
	return p.parseLines(lines), nil
}

func (p *TypeScriptParser) parseLines(lines []string) []model.CodeEntity {
	var entities []model.CodeEntity

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if match := p.classRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			sig := "class " + name
			if match[2] != "" {
				ext := strings.TrimSpace(match[2])
				if ext != "" {
					sig += " " + ext
				}
			}

			docstring := p.extractTSDoc(lines, i-1)
			methods := p.extractClassMethods(lines, i+1)
			classIndent := p.getIndent(line)

			endLine := i + 1
			for j := i + 1; j < len(lines); j++ {
				if p.getIndent(lines[j]) <= classIndent && strings.TrimSpace(lines[j]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[j]), "//") && !strings.HasPrefix(strings.TrimSpace(lines[j]), "*") {
					break
				}
				endLine = j + 1
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityClass,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
				LineEnd:   endLine,
				Children:  methods,
			})
		} else if match := p.interfaceRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			sig := "interface " + name
			if match[2] != "" {
				sig += " extends " + strings.TrimSpace(match[2])
			}

			docstring := p.extractTSDoc(lines, i-1)
			properties := p.extractInterfaceProperties(lines, i+1)
			if properties != "" {
				sig += " { " + properties + " }"
			}

			endLine := p.findBlockEnd(lines, i)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityInterface,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
				LineEnd:   endLine,
			})
		} else if match := p.typeRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			docstring := p.extractTSDoc(lines, i-1)
			typeBody := p.extractTypeBody(lines, i)
			sig := "type " + name + " = " + typeBody

			endLine := p.findBlockEnd(lines, i)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityStruct,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
				LineEnd:   endLine,
			})
		} else if match := p.enumRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			sig := "enum " + name
			docstring := p.extractTSDoc(lines, i-1)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityEnum,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		} else if match := p.funcRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			params := match[2]
			returnType := ""
			if len(match) > 3 {
				returnType = strings.TrimSpace(match[3])
			}

			sig := "function " + name + "(" + params + ")"
			if returnType != "" {
				sig += ": " + returnType
			}

			docstring := p.extractTSDoc(lines, i-1)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityFunction,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		} else if match := p.arrowFuncRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			sig := "const " + name + " = (...) =>"
			docstring := p.extractTSDoc(lines, i-1)

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

func (p *TypeScriptParser) extractClassMethods(lines []string, startIdx int) []model.CodeEntity {
	var methods []model.CodeEntity
	classIndent := -1

	if startIdx < len(lines) {
		classIndent = p.getIndent(lines[startIdx-1])
	}

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || trimmed == "{" || trimmed == "}" {
			continue
		}

		currentIndent := p.getIndent(line)
		if currentIndent <= classIndent && trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "*") {
			break
		}

		if match := p.methodRegex.FindStringSubmatch(line); match != nil {
			methodName := match[1]
			if p.isValidMethodName(methodName) && currentIndent > classIndent {
				params := match[2]
				returnType := ""
				if len(match) > 3 {
					returnType = strings.TrimSpace(match[3])
				}

				methodSig := methodName + "(" + params + ")"
				if returnType != "" {
					methodSig += ": " + returnType
				}

				methodDoc := p.extractTSDoc(lines, i-1)

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

	return methods
}

func (p *TypeScriptParser) getIndent(line string) int {
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

func (p *TypeScriptParser) isValidMethodName(name string) bool {
	keywords := map[string]bool{
		"if": true, "else": true, "for": true, "while": true,
		"return": true, "const": true, "let": true, "var": true,
		"function": true, "class": true, "export": true, "import": true,
		"interface": true, "type": true, "enum": true, "namespace": true,
		"module": true, "declare": true, "abstract": true,
		"public": true, "private": true, "protected": true, "readonly": true,
		"static": true, "new": true,
		"super": true, "this": true,
	}
	return !keywords[name] && !strings.HasPrefix(name, "//")
}

func (p *TypeScriptParser) extractTSDoc(lines []string, lineIdx int) string {
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

func (p *TypeScriptParser) extractTypeBody(lines []string, startIdx int) string {
	if startIdx >= len(lines) {
		return ""
	}

	line := lines[startIdx]
	trimmed := strings.TrimSpace(line)

	eqIdx := strings.Index(trimmed, "=")
	if eqIdx == -1 {
		return ""
	}

	afterEq := strings.TrimSpace(trimmed[eqIdx+1:])
	if afterEq == "" {
		return ""
	}

	if strings.HasPrefix(afterEq, "{") {
		return p.extractBlockContent(lines, startIdx, afterEq)
	}

	if strings.HasSuffix(afterEq, ";") {
		return strings.TrimSuffix(afterEq, ";")
	}
	return afterEq
}

func (p *TypeScriptParser) extractInterfaceProperties(lines []string, startIdx int) string {
	if startIdx >= len(lines) {
		return ""
	}

	var props []string
	braceDepth := 0
	started := false

	if startIdx > 0 {
		prevLine := lines[startIdx-1]
		trimmed := strings.TrimSpace(prevLine)
		if strings.Contains(trimmed, "{") {
			started = true
			braceDepth = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if braceDepth == 0 {
				return ""
			}
		}
	}

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if !started {
			if strings.Contains(trimmed, "{") {
				started = true
				braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
				if braceDepth == 0 {
					break
				}
			}
			continue
		}

		braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
		if braceDepth == 0 {
			break
		}

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		propMatch := regexp.MustCompile(`^(\w+)(\?)?\s*:\s*([^;]+);?$`).FindStringSubmatch(trimmed)
		if propMatch != nil {
			propName := propMatch[1]
			propType := strings.TrimSuffix(propMatch[3], ";")
			propType = strings.TrimSpace(propType)
			if propMatch[2] == "?" {
				props = append(props, propName+"?: "+propType)
			} else {
				props = append(props, propName+": "+propType)
			}
		}
	}

	return strings.Join(props, ", ")
}

func (p *TypeScriptParser) extractBlockContent(lines []string, startIdx int, firstLine string) string {
	var content strings.Builder
	content.WriteString(firstLine)

	braceDepth := strings.Count(firstLine, "{") - strings.Count(firstLine, "}")

	if braceDepth == 0 {
		return strings.TrimSuffix(firstLine, ";")
	}

	for i := startIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		content.WriteString(" ")
		content.WriteString(trimmed)

		braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")

		if braceDepth == 0 {
			break
		}
	}

	result := content.String()
	result = strings.ReplaceAll(result, "{ ", "{")
	result = strings.ReplaceAll(result, " }", "}")
	result = strings.ReplaceAll(result, "; ", "; ")
	return result
}

func (p *TypeScriptParser) findBlockEnd(lines []string, startIdx int) int {
	if startIdx >= len(lines) {
		return startIdx + 1
	}

	startIndent := p.getIndent(lines[startIdx])

	for i := startIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}

		currentIndent := p.getIndent(lines[i])
		if currentIndent <= startIndent && trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "*") {
			return i
		}
	}

	return len(lines)
}

func init() {
	Register(NewTypeScriptParser())
}
