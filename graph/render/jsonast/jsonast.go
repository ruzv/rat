package jsonast

// AstAttributes describes a abstract syntax tree part attributes.
type AstAttributes map[string]any

// AstPart describes a abstract syntax tree part.
type AstPart struct {
	Type       string        `json:"type"`
	Attributes AstAttributes `json:"attributes,omitempty"`
	Children   []*AstPart    `json:"children,omitempty"`
	parent     *AstPart
}

// NewRootAstPart creates a new root ast part, setting the parent to itself.
func NewRootAstPart() *AstPart {
	rootPart := &AstPart{}
	rootPart.parent = rootPart

	return rootPart
}

// AddLeaf adds a leaf (a ast part that can not contain other ast parts) to
// the ast part.
func (part *AstPart) AddLeaf(leaf *AstPart) {
	leaf.parent = part
	part.Children = append(part.Children, leaf)
}

// AddContainer on entering adds a container (a ast part that can contain other
// ast parts) to the ast part, on exit moves the target part back one parent.
func (part *AstPart) AddContainer(child *AstPart, entering bool) *AstPart {
	if !entering { // exiting
		return part.parent
	}

	child.parent = part
	part.Children = append(part.Children, child)

	return child
}