package ignored

type node interface {
	rollback() int
	addChild(newChild node) node
	getAncestor(ancestor string) node
}

func newNode(oldNode node, path string, parent string, idx int, isDir bool) node {
	p := oldNode.getAncestor(parent)
	var n node
	if isDir {
		n = &dirNode{path, p, nil, idx}
	} else {
		n = &fileNode{p}
	}
	p.addChild(n)
	return n
}

type dirNode struct {
	path   string
	parent node
	child  node
	idx    int
}

func (n *dirNode) rollback() int               { return n.idx }
func (n *dirNode) addChild(newChild node) node { n.child = newChild; return n }

func (n *dirNode) getAncestor(ancestor string) node {
	if n.path == ancestor {
		return n
	} else {
		return n.parent.getAncestor(ancestor)
	}
}

type fileNode struct {
	parent node
}

func (n *fileNode) rollback() int                    { return n.parent.rollback() }
func (n *fileNode) addChild(newChild node) node      { return n }
func (n *fileNode) getAncestor(ancestor string) node { return n.parent.getAncestor(ancestor) }
