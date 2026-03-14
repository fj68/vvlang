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
	"github.com/fj68/vvlang/docstring"
	"github.com/fj68/vvlang/parser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldparser "github.com/yuin/goldmark/parser"
)

//go:embed doc-templates/*.html
var templatesFS embed.FS

//go:embed doc-static/*
var staticFS embed.FS

var (
	moduleTemplate *template.Template
	indexTemplate  *template.Template
	md             goldmark.Markdown
)

func init() {
	layout := template.Must(template.ParseFS(templatesFS, "doc-templates/layout.html"))

	moduleTemplate = template.Must(layout.Clone())
	template.Must(moduleTemplate.ParseFS(templatesFS, "doc-templates/module.html"))

	indexTemplate = template.Must(layout.Clone())
	template.Must(indexTemplate.ParseFS(templatesFS, "doc-templates/index.html"))

	md = goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.TaskList,
			extension.Strikethrough,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			goldparser.WithAutoHeadingID(),
		),
	)
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
	Signature string
	ExternTag string
	Name      string
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

	// Copy static files
	staticDir := filepath.Join(outputDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("mkdir static: %v", err)
	}

	entries, err := staticFS.ReadDir("doc-static")
	if err != nil {
		return fmt.Errorf("read static fs: %v", err)
	}

	for _, entry := range entries {
		data, err := staticFS.ReadFile("doc-static/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read static file %s: %v", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(staticDir, entry.Name()), data, 0644); err != nil {
			return fmt.Errorf("write static file %s: %v", entry.Name(), err)
		}
	}

	for _, mod := range modules {
		if doc := docstring.GetDocstring(mod); doc != nil {
			for lang := range doc {
				langs[lang] = true
			}
		}
		for _, stmt := range mod.Statements {
			if decl, ok := stmt.(*ast.VarDeclStmt); ok {
				if doc := docstring.GetDocstring(decl); doc != nil {
					for lang := range doc {
						langs[lang] = true
					}
				}
			} else if ext, ok := stmt.(*ast.ExternStmt); ok {
				if doc := docstring.GetDocstring(ext); doc != nil {
					for lang := range doc {
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
				Docstring:   template.HTML(formatDoc(getDocstring(docstring.GetDocstring(mod), lang))),
				AllLangs:    sortedLangs,
				CurrentLang: lang,
				RootRel:     rootRel,
			}

			for _, stmt := range mod.Statements {
				switch s := stmt.(type) {
				case *ast.VarDeclStmt:
					if !s.Exported {
						continue
					}
					exp := exportData{Name: s.Name}
					if fun, ok := s.Body.(*ast.FunLiteralExpr); ok {
						args := strings.Join(fun.Args, ", ")
						exp.Signature = fmt.Sprintf("fun %s(%s)", s.Name, args)
					} else {
						exp.Signature = fmt.Sprintf("let %s", s.Name)
					}
					exp.Doc = template.HTML(formatDoc(getDocstring(docstring.GetDocstring(s), lang)))
					data.Exports = append(data.Exports, exp)
				case *ast.ExternStmt:
					if !s.Exported {
						continue
					}
					exp := exportData{Name: s.Name}
					if s.Kind == "fun" {
						args := strings.Join(s.Args, ", ")
						exp.Signature = fmt.Sprintf("fun %s(%s)", s.Name, args)
					} else {
						exp.Signature = fmt.Sprintf("let %s", s.Name)
					}
					exp.ExternTag = fmt.Sprintf("extern %q", s.Type)
					exp.Doc = template.HTML(formatDoc(getDocstring(docstring.GetDocstring(s), lang)))
					data.Exports = append(data.Exports, exp)
				case *ast.RecFunDeclStmt:
					for _, fun := range s.Funs {
						if !fun.Exported {
							continue
						}
						exp := exportData{Name: fun.Name}
						if f, ok := fun.Body.(*ast.FunLiteralExpr); ok {
							args := strings.Join(f.Args, ", ")
							exp.Signature = fmt.Sprintf("fun %s(%s)", fun.Name, args)
						} else {
							exp.Signature = fmt.Sprintf("let %s", fun.Name)
						}
						exp.Doc = template.HTML(formatDoc(getDocstring(docstring.GetDocstring(fun), lang)))
						data.Exports = append(data.Exports, exp)
					}
				}
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
			modDoc := getDocstring(docstring.GetDocstring(mod), lang)
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
	var buf strings.Builder
	if err := md.Convert([]byte(doc), &buf); err != nil {
		return doc
	}
	return buf.String()
}
