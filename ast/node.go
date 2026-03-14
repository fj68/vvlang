package ast

type Node interface {
	Inspect() string
}

type Annotation interface {
	ApplyTo(node Node)
}

type AnnotatedNode interface {
	Node
	AddAnnotation(annotation Annotation)
	GetAnnotations() []Annotation
}
