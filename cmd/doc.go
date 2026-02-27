package cmd

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fj68/vvlang/ast"
	"github.com/fj68/vvlang/parser"
)

//go:embed doc-templates/*.html
var templatesFS embed.FS

var (
	moduleTemplate *template.Template
	indexTemplate  *template.Template
)

func init() {
	layout := template.Must(template.ParseFS(templatesFS, "doc-templates/layout.html"))

	moduleTemplate = template.Must(layout.Clone())
	template.Must(moduleTemplate.ParseFS(templatesFS, "doc-templates/module.html"))

	indexTemplate = template.Must(layout.Clone())
	template.Must(indexTemplate.ParseFS(templatesFS, "doc-templates/index.html"))
}

func Doc() {
	target := "."
	outputDir := "docs"

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--output" || arg == "-o" {
			if i+1 < len(os.Args) {
				outputDir = os.Args[i+1]
				i++
			} else {
				fmt.Println("error: --output/-o requires an argument")
				os.Exit(1)
			}
		} else if strings.HasPrefix(arg, "-") {
			fmt.Printf("error: unknown option %s\n", arg)
			os.Exit(1)
		} else {
			target = arg
		}
	}

	files, err := collectVVFiles(target)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("no .vv files found")
		return
	}

	modules := make(map[string]*ast.Module)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("warning: could not read %s: %v\n", file, err)
			continue
		}

		mod, err := parser.Parse([]rune(string(content)))
		if err != nil {
			fmt.Printf("warning: could not parse %s: %v\n", file, err)
			continue
		}
		modules[file] = mod
	}

	rootDirName := target
	if target == "." {
		cwd, err := os.Getwd()
		if err == nil {
			rootDirName = filepath.Base(cwd)
		}
	} else {
		targetAbs, err := filepath.Abs(target)
		if err == nil {
			rootDirName = filepath.Base(targetAbs)
		}
	}

	if err := generateDocs(modules, outputDir, rootDirName); err != nil {
		fmt.Printf("error generating docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Documentation generated in %s\n", outputDir)
}

func collectVVFiles(target string) ([]string, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	var files []string
	if info.IsDir() {
		err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				name := info.Name()
				if name == ".vv-modules" || name == "vendor" || name == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) == ".vv" {
				files = append(files, path)
			}
			return nil
		})
		return files, err
	}

	if filepath.Ext(target) == ".vv" {
		return []string{target}, nil
	}
	return nil, fmt.Errorf("%s is not a .vv file", target)
}

type exportData struct {
	Kind      string
	Name      string
	Arguments string
	Doc       template.HTML
}

type modulePageData struct {
	ModName     string
	Docstring   template.HTML
	Exports     []exportData
	AllLangs    []string
	CurrentLang string
	RootRel     string
}

type indexPageData struct {
	Categories  []categoryData
	AllLangs    []string
	CurrentLang string
	RootRel     string
}

type categoryData struct {
	Name    string
	Modules []moduleMeta
}

type moduleMeta struct {
	Name    string
	Path    string
	Summary string
}

func generateDocs(modules map[string]*ast.Module, outputDir string, rootDirName string) error {
	langs := make(map[string]bool)
	langs["en"] = true

	for _, mod := range modules {
		if mod.Docstring != nil {
			for lang := range mod.Docstring {
				langs[lang] = true
			}
		}
		for _, stmt := range mod.Statements {
			if decl, ok := stmt.(*ast.VarDeclStmt); ok {
				if decl.Docstring != nil {
					for lang := range decl.Docstring {
						langs[lang] = true
					}
				}
			} else if ext, ok := stmt.(*ast.ExternStmt); ok {
				if ext.Docstring != nil {
					for lang := range ext.Docstring {
						langs[lang] = true
					}
				}
			}
		}
	}

	sortedLangs := []string{}
	for lang := range langs {
		sortedLangs = append(sortedLangs, lang)
	}

	for lang := range langs {
		langDir := filepath.Join(outputDir, lang)
		if err := os.MkdirAll(langDir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %v", langDir, err)
		}

		for path, mod := range modules {
			relPath, _ := filepath.Rel(".", path)
			modPath := filepath.ToSlash(relPath)

			modDir := filepath.Join(langDir, filepath.FromSlash(modPath))
			if err := os.MkdirAll(modDir, 0755); err != nil {
				return err
			}

			depth := strings.Count(modPath, "/") + 2
			rootRel := strings.Repeat("../", depth)

			data := modulePageData{
				ModName:     modPath,
				Docstring:   template.HTML(formatDoc(getDocstring(mod.Docstring, lang))),
				AllLangs:    sortedLangs,
				CurrentLang: lang,
				RootRel:     rootRel,
			}

			for name, stmt := range mod.Exports {
				exp := exportData{Name: name}
				switch s := stmt.(type) {
				case *ast.VarDeclStmt:
					exp.Kind = "variable"
					if fun, ok := s.Body.(*ast.FunLiteralExpr); ok {
						exp.Kind = "function"
						exp.Arguments = "(" + strings.Join(fun.Args, ", ") + ")"
					}
					exp.Doc = template.HTML(formatDoc(getDocstring(s.Docstring, lang)))
				case *ast.ExternStmt:
					exp.Kind = "extern"
					exp.Doc = template.HTML(formatDoc(getDocstring(s.Docstring, lang)))
				}
				data.Exports = append(data.Exports, exp)
			}

			f, err := os.Create(filepath.Join(modDir, "index.html"))
			if err != nil {
				return err
			}
			if err := moduleTemplate.ExecuteTemplate(f, "layout", data); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}

		indexData := indexPageData{
			AllLangs:    sortedLangs,
			CurrentLang: lang,
			RootRel:     "../",
		}

		// Group modules by directory
		catMap := make(map[string][]moduleMeta)
		for path, mod := range modules {
			relPath, _ := filepath.Rel(".", path)
			modPath := filepath.ToSlash(relPath)

			cat := rootDirName
			if idx := strings.LastIndex(modPath, "/"); idx != -1 {
				cat = modPath[:idx]
			}

			summary := ""
			modDoc := getDocstring(mod.Docstring, lang)
			if modDoc != "" {
				lines := strings.Split(modDoc, "\n")
				if len(lines) > 0 {
					summary = lines[0]
				}
			}

			catMap[cat] = append(catMap[cat], moduleMeta{
				Name:    modPath,
				Path:    modPath,
				Summary: summary,
			})
		}

		// Convert map to sorted slice
		for catName, mods := range catMap {
			sort.Slice(mods, func(i, j int) bool {
				return mods[i].Name < mods[j].Name
			})
			indexData.Categories = append(indexData.Categories, categoryData{
				Name:    catName,
				Modules: mods,
			})
		}
		sort.Slice(indexData.Categories, func(i, j int) bool {
			return indexData.Categories[i].Name < indexData.Categories[j].Name
		})

		f, err := os.Create(filepath.Join(langDir, "index.html"))
		if err != nil {
			return err
		}
		if err := indexTemplate.ExecuteTemplate(f, "layout", indexData); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

func getDocstring(docs map[string]string, lang string) string {
	if val, ok := docs[lang]; ok && val != "" {
		return val
	}
	return docs["en"]
}

func formatDoc(doc string) string {
	lines := strings.Split(doc, "\n")
	return strings.Join(lines, "<br>\n")
}
