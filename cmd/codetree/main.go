package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lemantorus/codetree/internal/model"
	"github.com/lemantorus/codetree/internal/output"
	"github.com/lemantorus/codetree/internal/parser"
	"github.com/lemantorus/codetree/internal/tree"

	_ "github.com/lemantorus/codetree/internal/parser"

	"github.com/spf13/cobra"
)

var (
	depth             int
	languages         []string
	entityTypes       []string
	noDocstring       bool
	noSignatures      bool
	outputFile        string
	extensions        []string
	includeLibs       bool
	excludePattern    string
	excludeIgnoreCase bool
	maxSignatureLen   int
	maxDocstringLen   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "codetree [directory]",
		Short: "Build repository tree with code signatures",
		Args:  cobra.MaximumNArgs(1),
		RunE:  run,
	}

	rootCmd.Flags().IntVarP(&depth, "depth", "d", -1, "max search depth (-1 = unlimited)")
	rootCmd.Flags().StringSliceVarP(&languages, "lang", "l", []string{}, "languages to parse (comma-separated)")
	rootCmd.Flags().StringSliceVarP(&entityTypes, "type", "t", []string{}, "entity types: func,class,method,interface,enum,type,const,var")
	rootCmd.Flags().BoolVar(&noDocstring, "no-docstrings", false, "hide docstrings")
	rootCmd.Flags().BoolVar(&noSignatures, "no-signatures", false, "hide function/class signatures (show only type + name)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (default: stdout)")
	rootCmd.Flags().StringSliceVar(&extensions, "ext", []string{}, "file extensions (default: auto by language)")
	rootCmd.Flags().BoolVar(&includeLibs, "include-libs", false, "include dependency directories (venv, node_modules, etc.)")
	rootCmd.Flags().StringVar(&excludePattern, "exclude-dirs", "", "regex pattern for directories to skip (use | as separator for multiple patterns)")
	rootCmd.Flags().BoolVar(&excludeIgnoreCase, "exclude-dirs-ignore-case", false, "case-insensitive matching for --exclude-dirs")
	rootCmd.Flags().IntVar(&maxSignatureLen, "max-signature", 0, "max signature length (0 = unlimited)")
	rootCmd.Flags().IntVar(&maxDocstringLen, "max-docstring", 0, "max docstring length (0 = unlimited)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	entityTypeMap := parseEntityTypes(entityTypes)
	if len(languages) > 0 {
		validateLanguages(languages)
	}

	opts := []tree.BuilderOption{
		tree.WithMaxDepth(depth),
		tree.WithDocstrings(!noDocstring),
		tree.WithIncludeLibs(includeLibs),
		tree.WithExcludePattern(excludePattern, excludeIgnoreCase),
	}
	if len(languages) > 0 {
		opts = append(opts, tree.WithLanguages(languages))
	}
	if len(entityTypeMap) > 0 {
		types := make([]model.EntityType, 0, len(entityTypeMap))
		for t := range entityTypeMap {
			types = append(types, t)
		}
		opts = append(opts, tree.WithEntityTypes(types))
	}
	if len(extensions) > 0 {
		opts = append(opts, tree.WithExtensions(extensions))
	}

	builder := tree.NewBuilder(opts...)
	root, err := builder.Build(dir)
	if err != nil {
		return fmt.Errorf("failed to build tree: %w", err)
	}

	formatter := output.NewFormatter(!noDocstring, !noSignatures, maxSignatureLen, maxDocstringLen)

	var out *os.File
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	return formatter.Format(root, out)
}

func parseEntityTypes(types []string) map[model.EntityType]bool {
	result := make(map[model.EntityType]bool)
	for _, t := range types {
		switch strings.ToLower(t) {
		case "func", "function":
			result[model.EntityFunction] = true
		case "class":
			result[model.EntityClass] = true
		case "method":
			result[model.EntityMethod] = true
		case "const", "constant":
			result[model.EntityConstant] = true
		case "var", "variable":
			result[model.EntityVariable] = true
		case "interface":
			result[model.EntityInterface] = true
		case "struct", "type":
			result[model.EntityStruct] = true
		case "enum":
			result[model.EntityEnum] = true
		}
	}
	return result
}

func validateLanguages(langs []string) {
	available := parser.AvailableLanguages()
	availableMap := make(map[string]bool)
	for _, lang := range available {
		availableMap[lang] = true
	}

	for _, lang := range langs {
		if !availableMap[lang] {
			fmt.Fprintf(os.Stderr, "Warning: language '%s' is not supported. Available: %v\n", lang, available)
		}
	}
}
