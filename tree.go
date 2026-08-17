package ignored

type node interface {
	getPath() string
	getParent() node
	rollback() int
	setRollback(idx int)
	addChild(newChild node) node
	getAncestor(ancestor string) node
}

func newNode(oldNode node, path string, parent string, isDir bool) node {
	p := oldNode.getAncestor(parent)
	var n node
	if isDir {
		n = &dirNode{path, p, nil, 0}
	} else {
		n = &fileNode{path, p}
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

func (n *dirNode) getPath() string             { return n.path }
func (n *dirNode) rollback() int               { return n.idx }
func (n *dirNode) setRollback(idx int)         { n.idx = idx }
func (n *dirNode) addChild(newChild node) node { n.child = newChild; return n }
func (n *dirNode) getParent() node             { return n.parent }

func (n *dirNode) getAncestor(ancestor string) node {
	curr := node(n)
	for curr != nil {
		if curr.getPath() == ancestor {
			return curr
		}
		curr = curr.getParent()
	}
	return nil
}

type fileNode struct {
	path   string
	parent node
}

func (n *fileNode) getPath() string { return n.path }
func (n *fileNode) getParent() node { return n.parent }

func (n *fileNode) rollback() int {
	curr := n.parent
	return curr.rollback()
}

func (*fileNode) setRollback(_ int)      {}
func (n *fileNode) addChild(_ node) node { return n }

func (n *fileNode) getAncestor(ancestor string) node {
	curr := n.parent
	for curr != nil {
		if curr.getPath() == ancestor {
			return curr
		}
		curr = curr.getParent()
	}
	return nil
}
