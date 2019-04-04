package galaxy

import (
	"fmt"
	"github.com/xlab/treeprint"
)

type Printer struct {
	context *Context
}

func (p *Printer) Tree() {
	tree := treeprint.New()
	for ns, files := range p.context.GetNamespaceFilesMap() {
		branch := tree.AddBranch(ns)
		for _, file := range files {
			branch.AddNode(file)
		}
	}
	fmt.Println(tree.String())
}

func NewPrinter(context *Context) *Printer {
	return &Printer{context: context}
}
