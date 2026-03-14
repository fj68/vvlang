package docstring

import "github.com/fj68/vvlang/ast"

type DocstringAnnotation struct {
	Languages map[string]string
	Position  *ast.Pos
}

func (d *DocstringAnnotation) ApplyTo(node ast.Node) {
	if annotated, ok := node.(ast.AnnotatedNode); ok {
		annotated.AddAnnotation(d)
	}
}

func GetDocstring(node ast.Node) map[string]string {
	if an, ok := node.(ast.AnnotatedNode); ok {
		for _, a := range an.GetAnnotations() {
			if doc, ok := a.(*DocstringAnnotation); ok {
				return doc.Languages
			}
		}
	}
	return nil
}
