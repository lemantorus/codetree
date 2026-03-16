package output

import (
	"fmt"
	"io"
	"strings"

	"codetree/internal/model"
)

type Formatter struct {
	showDocstrings  bool
	showSignatures  bool
	maxSignatureLen int
	maxDocstringLen int
	indent          string
}

func NewFormatter(showDocstrings bool, showSignatures bool, maxSignatureLen int, maxDocstringLen int) *Formatter {
	return &Formatter{
		showDocstrings:  showDocstrings,
		showSignatures:  showSignatures,
		maxSignatureLen: maxSignatureLen,
		maxDocstringLen: maxDocstringLen,
		indent:          "    ",
	}
}

func (f *Formatter) Format(root *model.DirNode, w io.Writer) error {
	fmt.Fprintln(w, root.Name+"/")
	return f.formatNode(root, w, "")
}

func (f *Formatter) formatNode(node *model.DirNode, w io.Writer, prefix string) error {
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if child.IsDir {
			fmt.Fprintf(w, "%s%s%s/\n", prefix, connector, child.Name)
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			f.formatNode(child, w, newPrefix)
		} else {
			fmt.Fprintf(w, "%s%s%s\n", prefix, connector, child.Name)
			if len(child.Entities) > 0 {
				newPrefix := prefix
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
				f.formatEntities(child.Entities, w, newPrefix, true)
			}
		}
	}
	return nil
}

func (f *Formatter) formatEntities(entities []model.CodeEntity, w io.Writer, prefix string, isFileLevel bool) {
	for i, entity := range entities {
		isLast := i == len(entities)-1 && len(entity.Children) == 0
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		signature := entity.Signature
		if f.showSignatures && signature != "" {
			signature = f.truncate(signature, f.maxSignatureLen)
			fmt.Fprintf(w, "%s%s%s\n", prefix, connector, signature)
		} else {
			fmt.Fprintf(w, "%s%s%s %s\n", prefix, connector, entity.Type, entity.Name)
		}

		entityPrefix := prefix
		if isLast {
			entityPrefix += "    "
		} else {
			entityPrefix += "│   "
		}

		if f.showDocstrings && entity.Docstring != "" {
			docstring := f.truncate(entity.Docstring, f.maxDocstringLen)
			docLines := strings.Split(docstring, "\n")
			for _, line := range docLines {
				fmt.Fprintf(w, "%s│   %s\n", entityPrefix, line)
			}
		}

		if len(entity.Children) > 0 {
			childPrefix := entityPrefix
			f.formatEntities(entity.Children, w, childPrefix, false)
		}
	}
}

func (f *Formatter) truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func FormatTree(root *model.DirNode, showDocstrings bool, showSignatures bool, maxSignatureLen int, maxDocstringLen int) string {
	var sb strings.Builder
	f := NewFormatter(showDocstrings, showSignatures, maxSignatureLen, maxDocstringLen)
	f.Format(root, &sb)
	return sb.String()
}
