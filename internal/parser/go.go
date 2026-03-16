package parser

import (
	"regexp"
	"strings"

	"github.com/lemantorus/codetree/internal/model"
)

type GoParser struct {
	funcRegex      *regexp.Regexp
	methodRegex    *regexp.Regexp
	structRegex    *regexp.Regexp
	interfaceRegex *regexp.Regexp
	typeRegex      *regexp.Regexp
	constRegex     *regexp.Regexp
	constBlockRe   *regexp.Regexp
}

func NewGoParser() *GoParser {
	return &GoParser{
		funcRegex:      regexp.MustCompile(`(?m)^func\s+([A-Za-z]\w*)\s*\(([^)]*)\)(?:\s*([^\{]+))?`),
		methodRegex:    regexp.MustCompile(`(?m)^func\s+\(([^)]+)\)\s*([A-Za-z]\w*)\s*\(([^)]*)\)(?:\s*([^\{]+))?`),
		structRegex:    regexp.MustCompile(`(?m)^type\s+([A-Za-z]\w*)\s+struct\b`),
		interfaceRegex: regexp.MustCompile(`(?m)^type\s+([A-Za-z]\w*)\s+interface\b`),
		typeRegex:      regexp.MustCompile(`(?m)^type\s+([A-Za-z]\w*)\s+(.+)`),
		constRegex:     regexp.MustCompile(`(?m)^\s*const\s+(\w+)\s*(?:=\s*(.+))?`),
		constBlockRe:   regexp.MustCompile(`(?m)^\s*(\w+)\s*(?:=\s*(.+))?`),
	}
}

func (p *GoParser) Language() string {
	return "go"
}

func (p *GoParser) Extensions() []string {
	return []string{".go"}
}

func (p *GoParser) LibDirs() []string {
	return []string{
		"vendor",
		"Godeps",
		".godeps",
	}
}

func (p *GoParser) Parse(content []byte) ([]model.CodeEntity, error) {
	lines := strings.Split(string(content), "\n")
	return p.parseLines(lines), nil
}

func (p *GoParser) parseLines(lines []string) []model.CodeEntity {
	var entities []model.CodeEntity
	processedMethods := make(map[int]bool)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || trimmed == "" {
			continue
		}

		if match := p.structRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			sig := "type " + name + " struct"
			docstring := p.extractGoDoc(lines, i-1)
			fields := p.extractStructFields(lines, i+1)
			endLine := p.findBlockEnd(lines, i)

			if len(fields) > 0 {
				sig += " { " + strings.Join(fields, ", ") + " }"
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityStruct,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
				LineEnd:   endLine,
			})
		} else if match := p.interfaceRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			sig := "type " + name + " interface"
			docstring := p.extractGoDoc(lines, i-1)
			methods := p.extractInterfaceMethods(lines, i+1)
			endLine := p.findBlockEnd(lines, i)

			if len(methods) > 0 {
				sig += " { " + strings.Join(methods, ", ") + " }"
			}

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityInterface,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
				LineEnd:   endLine,
			})
		} else if match := p.typeRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			typeDef := strings.TrimSpace(match[2])
			if strings.HasSuffix(typeDef, ";") {
				typeDef = strings.TrimSuffix(typeDef, ";")
			}
			sig := "type " + name + " " + typeDef
			docstring := p.extractGoDoc(lines, i-1)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityStruct,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		} else if match := p.methodRegex.FindStringSubmatch(line); match != nil {
			receiver := strings.TrimSpace(match[1])
			name := match[2]
			params := match[3]
			returnType := ""
			if len(match) > 4 {
				returnType = strings.TrimSpace(match[4])
			}
			returnType = strings.TrimSuffix(returnType, "{")

			sig := "func (" + receiver + ") " + name + "(" + params + ")"
			if returnType != "" {
				sig += " " + returnType
			}

			docstring := p.extractGoDoc(lines, i-1)
			processedMethods[i] = true

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityMethod,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		} else if match := p.funcRegex.FindStringSubmatch(line); match != nil {
			if processedMethods[i] {
				continue
			}
			name := match[1]
			params := match[2]
			returnType := ""
			if len(match) > 3 {
				returnType = strings.TrimSpace(match[3])
			}
			returnType = strings.TrimSuffix(returnType, "{")

			sig := "func " + name + "(" + params + ")"
			if returnType != "" {
				sig += " " + returnType
			}

			docstring := p.extractGoDoc(lines, i-1)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityFunction,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		} else if strings.HasPrefix(trimmed, "const (") {
			docstring := p.extractGoDoc(lines, i-1)
			consts := p.extractConstBlock(lines, i+1)
			for _, c := range consts {
				c.Docstring = docstring
				entities = append(entities, c)
			}
		} else if match := p.constRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			value := ""
			if len(match) > 2 && match[2] != "" {
				value = strings.TrimSpace(match[2])
			}
			sig := "const " + name
			if value != "" {
				sig += " = " + value
			}
			docstring := p.extractGoDoc(lines, i-1)

			entities = append(entities, model.CodeEntity{
				Name:      name,
				Type:      model.EntityConstant,
				Signature: sig,
				Docstring: docstring,
				LineStart: i + 1,
			})
		}
	}

	return entities
}

func (p *GoParser) extractGoDoc(lines []string, lineIdx int) string {
	if lineIdx < 0 {
		return ""
	}

	var docLines []string
	inBlockComment := false

	for i := lineIdx; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])

		if trimmed == "" {
			if len(docLines) > 0 {
				break
			}
			continue
		}

		if strings.HasPrefix(trimmed, "/*") {
			if strings.HasSuffix(trimmed, "*/") {
				doc := strings.TrimPrefix(trimmed, "/*")
				doc = strings.TrimSuffix(doc, "*/")
				doc = strings.TrimSpace(doc)
				if doc != "" {
					docLines = append([]string{doc}, docLines...)
				}
			} else {
				inBlockComment = true
				doc := strings.TrimPrefix(trimmed, "/*")
				doc = strings.TrimSpace(doc)
				if doc != "" {
					docLines = append([]string{doc}, docLines...)
				}
			}
			break
		}

		if inBlockComment {
			if strings.HasSuffix(trimmed, "*/") {
				doc := strings.TrimSuffix(trimmed, "*/")
				doc = strings.TrimSpace(doc)
				if doc != "" {
					docLines = append([]string{doc}, docLines...)
				}
				break
			}
			doc := strings.TrimSpace(trimmed)
			if doc != "" {
				docLines = append([]string{doc}, docLines...)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "//") {
			doc := strings.TrimPrefix(trimmed, "//")
			doc = strings.TrimSpace(doc)
			if doc != "" {
				docLines = append([]string{doc}, docLines...)
			}
			continue
		}

		break
	}

	return strings.Join(docLines, "\n")
}

func (p *GoParser) extractStructFields(lines []string, startIdx int) []string {
	var fields []string
	braceDepth := 0
	started := false

	for i := startIdx; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "{") {
			started = true
			braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if braceDepth == 0 {
				break
			}
			trimmed = strings.TrimPrefix(trimmed, "{")
			trimmed = strings.TrimSpace(trimmed)
		}

		if !started {
			if trimmed == "{" {
				started = true
				braceDepth = 1
				continue
			}
			break
		}

		braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
		if braceDepth <= 0 {
			break
		}

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		if strings.HasPrefix(trimmed, "}") {
			break
		}

		fieldMatch := regexp.MustCompile(`^(\w+)\s+(.+)`).FindStringSubmatch(trimmed)
		if fieldMatch != nil && !strings.HasPrefix(fieldMatch[1], " ") {
			fieldName := fieldMatch[1]
			fieldType := strings.TrimSuffix(fieldMatch[2], "}")
			fieldType = strings.TrimSpace(fieldType)
			if !strings.Contains(fieldName, ":") && !strings.HasPrefix(fieldName, "//") {
				fields = append(fields, fieldName+" "+fieldType)
			}
		}
	}

	return fields
}

func (p *GoParser) extractInterfaceMethods(lines []string, startIdx int) []string {
	var methods []string
	braceDepth := 0
	started := false

	for i := startIdx; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		if trimmed == "" {
			continue
		}

		if strings.Contains(trimmed, "{") {
			started = true
			braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if braceDepth == 0 {
				break
			}
			trimmed = strings.TrimPrefix(trimmed, "{")
			trimmed = strings.TrimSpace(trimmed)
		}

		if !started {
			if trimmed == "{" {
				started = true
				braceDepth = 1
				continue
			}
			break
		}

		braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
		if braceDepth <= 0 {
			break
		}

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		if strings.HasPrefix(trimmed, "}") {
			break
		}

		methodMatch := regexp.MustCompile(`^(\w+)\s*\(([^)]*)\)\s*(.*)$`).FindStringSubmatch(trimmed)
		if methodMatch != nil {
			methodName := methodMatch[1]
			params := methodMatch[2]
			returnType := strings.TrimSpace(methodMatch[3])
			returnType = strings.TrimSuffix(returnType, "}")
			returnType = strings.TrimSpace(returnType)
			sig := methodName + "(" + params + ")"
			if returnType != "" {
				sig += " " + returnType
			}
			methods = append(methods, sig)
		} else if !strings.Contains(trimmed, "embed") && !strings.HasPrefix(trimmed, "//") {
			if trimmed != "" && !strings.HasPrefix(trimmed, "}") {
				trimmed = strings.TrimSuffix(trimmed, "}")
				methods = append(methods, trimmed)
			}
		}
	}

	return methods
}

func (p *GoParser) extractConstBlock(lines []string, startIdx int) []model.CodeEntity {
	var consts []model.CodeEntity

	for i := startIdx; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if trimmed == ")" {
			break
		}

		if match := p.constBlockRe.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			value := ""
			if len(match) > 2 && match[2] != "" {
				value = strings.TrimSpace(match[2])
			}
			sig := "const " + name
			if value != "" {
				sig += " = " + value
			}

			consts = append(consts, model.CodeEntity{
				Name:      name,
				Type:      model.EntityConstant,
				Signature: sig,
				LineStart: i + 1,
			})
		}
	}

	return consts
}

func (p *GoParser) findBlockEnd(lines []string, startIdx int) int {
	braceDepth := 0
	started := false

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]

		if strings.Contains(line, "{") {
			started = true
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
		} else if started {
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
		}

		if started && braceDepth == 0 {
			return i + 1
		}
	}

	return len(lines)
}

func init() {
	Register(NewGoParser())
}
