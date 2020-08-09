package kbinxml

import "github.com/beevik/etree"

//IterWalker structure
type IterWalker struct {
	doc  *etree.Document
	node *nodeWalker
	eof  bool
}

type nodeWalker struct {
	node              *etree.Element
	currentChildIndex int
	parent            *nodeWalker
}

// IterWalk through a xml document
func IterWalk(doc *etree.Document) *IterWalker {
	return &IterWalker{doc, &nodeWalker{doc.Root(), 0, nil}, false}
}

// Walk a step
func (i *IterWalker) Walk() (node *etree.Element, event string) {
	if i.eof {
		return i.doc.Root(), "eof"
	}
	children := i.node.node.ChildElements()
	if i.node.currentChildIndex == len(children) {
		event = "end"
		retNode := i.node.node
		if i.node.parent == nil {
			i.eof = true
		} else {
			i.node.parent.currentChildIndex++
			i.node = i.node.parent
		}
		return retNode, event
	}
	newNode := &nodeWalker{children[i.node.currentChildIndex],
		0,
		i.node}
	i.node = newNode
	return i.node.node, "start"

}
