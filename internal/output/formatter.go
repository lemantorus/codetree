package output

import (
	"fmt"
	"io"
	"strings"

	"codetree/internal/model"
)

type Formatter struct {
	showDocstrings bool
	showSignatures bool
	indent         string
}

func NewFormatter(showDocstrings bool, showSignatures bool) *Formatter {
	return &Formatter{
		showDocstrings: showDocstrings,
		showSignatures: showSignatures,
		indent:         "    ",
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

		if f.showSignatures && entity.Signature != "" {
			fmt.Fprintf(w, "%s%s%s\n", prefix, connector, entity.Signature)
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
			docLines := strings.Split(entity.Docstring, "\n")
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

func FormatTree(root *model.DirNode, showDocstrings bool, showSignatures bool) string {
	var sb strings.Builder
	f := NewFormatter(showDocstrings, showSignatures)
	f.Format(root, &sb)
	return sb.String()
}
