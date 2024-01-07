package launder

import (
	"strings"

	"golang.org/x/net/html"
)

func (s *Selection) After(selector string) *Selection {
	return s.AfterMatcher(compileMatcher(selector))
}

func (s *Selection) AfterMatcher(m Matcher) *Selection {
	return s.AfterNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) AfterSelection(sel *Selection) *Selection {
	return s.AfterNodes(sel.Nodes...)
}

func (s *Selection) AfterHtml(htmlStr string) *Selection {
	return s.eachNodeHtml(htmlStr, true, func(node *html.Node, nodes []*html.Node) {
		nextSibling := node.NextSibling
		for _, n := range nodes {
			if node.Parent != nil {
				node.Parent.InsertBefore(n, nextSibling)
			}
		}
	})
}

func (s *Selection) AfterNodes(ns ...*html.Node) *Selection {
	return s.manipulateNodes(ns, true, func(sn *html.Node, n *html.Node) {
		if sn.Parent != nil {
			sn.Parent.InsertBefore(n, sn.NextSibling)
		}
	})
}

func (s *Selection) Append(selector string) *Selection {
	return s.AppendMatcher(compileMatcher(selector))
}

func (s *Selection) AppendMatcher(m Matcher) *Selection {
	return s.AppendNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) AppendSelection(sel *Selection) *Selection {
	return s.AppendNodes(sel.Nodes...)
}

func (s *Selection) AppendHtml(htmlStr string) *Selection {
	return s.eachNodeHtml(htmlStr, false, func(node *html.Node, nodes []*html.Node) {
		for _, n := range nodes {
			node.AppendChild(n)
		}
	})
}

func (s *Selection) AppendNodes(ns ...*html.Node) *Selection {
	return s.manipulateNodes(ns, false, func(sn *html.Node, n *html.Node) {
		sn.AppendChild(n)
	})
}

func (s *Selection) Before(selector string) *Selection {
	return s.BeforeMatcher(compileMatcher(selector))
}

func (s *Selection) BeforeMatcher(m Matcher) *Selection {
	return s.BeforeNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) BeforeSelection(sel *Selection) *Selection {
	return s.BeforeNodes(sel.Nodes...)
}

func (s *Selection) BeforeHtml(htmlStr string) *Selection {
	return s.eachNodeHtml(htmlStr, true, func(node *html.Node, nodes []*html.Node) {
		for _, n := range nodes {
			if node.Parent != nil {
				node.Parent.InsertBefore(n, node)
			}
		}
	})
}

func (s *Selection) BeforeNodes(ns ...*html.Node) *Selection {
	return s.manipulateNodes(ns, false, func(sn *html.Node, n *html.Node) {
		if sn.Parent != nil {
			sn.Parent.InsertBefore(n, sn)
		}
	})
}

func (s *Selection) Clone() *Selection {
	ns := newEmptySelection(s.document)
	ns.Nodes = cloneNodes(s.Nodes)
	return ns
}

func (s *Selection) Empty() *Selection {
	var nodes []*html.Node

	for _, n := range s.Nodes {
		for c := n.FirstChild; c != nil; c = n.FirstChild {
			n.RemoveChild(c)
			nodes = append(nodes, c)
		}
	}

	return pushStack(s, nodes)
}

func (s *Selection) Prepend(selector string) *Selection {
	return s.PrependMatcher(compileMatcher(selector))
}

func (s *Selection) PrependMatcher(m Matcher) *Selection {
	return s.PrependNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) PrependSelection(sel *Selection) *Selection {
	return s.PrependNodes(sel.Nodes...)
}

func (s *Selection) PrependHtml(htmlStr string) *Selection {
	return s.eachNodeHtml(htmlStr, false, func(node *html.Node, nodes []*html.Node) {
		firstChild := node.FirstChild
		for _, n := range nodes {
			node.InsertBefore(n, firstChild)
		}
	})
}

func (s *Selection) PrependNodes(ns ...*html.Node) *Selection {
	return s.manipulateNodes(ns, true, func(sn *html.Node, n *html.Node) {
		sn.InsertBefore(n, sn.FirstChild)
	})
}

func (s *Selection) Remove() *Selection {
	for _, n := range s.Nodes {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}

	return s
}

func (s *Selection) RemoveFiltered(selector string) *Selection {
	return s.RemoveMatcher(compileMatcher(selector))
}

func (s *Selection) RemoveMatcher(m Matcher) *Selection {
	return s.FilterMatcher(m).Remove()
}

func (s *Selection) ReplaceWith(selector string) *Selection {
	return s.ReplaceWithMatcher(compileMatcher(selector))
}

func (s *Selection) ReplaceWithMatcher(m Matcher) *Selection {
	return s.ReplaceWithNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) ReplaceWithSelection(sel *Selection) *Selection {
	return s.ReplaceWithNodes(sel.Nodes...)
}

func (s *Selection) ReplaceWithHtml(htmlStr string) *Selection {
	s.eachNodeHtml(htmlStr, true, func(node *html.Node, nodes []*html.Node) {
		nextSibling := node.NextSibling
		for _, n := range nodes {
			if node.Parent != nil {
				node.Parent.InsertBefore(n, nextSibling)
			}
		}
	})
	return s.Remove()
}

func (s *Selection) ReplaceWithNodes(ns ...*html.Node) *Selection {
	s.AfterNodes(ns...)
	return s.Remove()
}

func (s *Selection) SetHtml(htmlStr string) *Selection {
	for _, context := range s.Nodes {
		for c := context.FirstChild; c != nil; c = context.FirstChild {
			context.RemoveChild(c)
		}
	}
	return s.eachNodeHtml(htmlStr, false, func(node *html.Node, nodes []*html.Node) {
		for _, n := range nodes {
			node.AppendChild(n)
		}
	})
}

func (s *Selection) SetText(text string) *Selection {
	return s.SetHtml(html.EscapeString(text))
}

func (s *Selection) Unwrap() *Selection {
	s.Parent().Each(func(i int, ss *Selection) {
		if ss.Nodes[0].Data != "body" {
			ss.ReplaceWithSelection(ss.Contents())
		}
	})

	return s
}

func (s *Selection) Wrap(selector string) *Selection {
	return s.WrapMatcher(compileMatcher(selector))
}

func (s *Selection) WrapMatcher(m Matcher) *Selection {
	return s.wrapNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) WrapSelection(sel *Selection) *Selection {
	return s.wrapNodes(sel.Nodes...)
}

func (s *Selection) WrapHtml(htmlStr string) *Selection {
	nodesMap := make(map[string][]*html.Node)
	for _, context := range s.Nodes {
		var parent *html.Node
		if context.Parent != nil {
			parent = context.Parent
		} else {
			parent = &html.Node{Type: html.ElementNode}
		}
		nodes, found := nodesMap[nodeName(parent)]
		if !found {
			nodes = parseHtmlWithContext(htmlStr, parent)
			nodesMap[nodeName(parent)] = nodes
		}
		newSingleSelection(context, s.document).wrapAllNodes(cloneNodes(nodes)...)
	}
	return s
}

func (s *Selection) WrapNode(n *html.Node) *Selection {
	return s.wrapNodes(n)
}

func (s *Selection) wrapNodes(ns ...*html.Node) *Selection {
	s.Each(func(i int, ss *Selection) {
		ss.wrapAllNodes(ns...)
	})

	return s
}

func (s *Selection) WrapAll(selector string) *Selection {
	return s.WrapAllMatcher(compileMatcher(selector))
}

func (s *Selection) WrapAllMatcher(m Matcher) *Selection {
	return s.wrapAllNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) WrapAllSelection(sel *Selection) *Selection {
	return s.wrapAllNodes(sel.Nodes...)
}

func (s *Selection) WrapAllHtml(htmlStr string) *Selection {
	var context *html.Node
	var nodes []*html.Node
	if len(s.Nodes) > 0 {
		context = s.Nodes[0]
		if context.Parent != nil {
			nodes = parseHtmlWithContext(htmlStr, context)
		} else {
			nodes = parseHtml(htmlStr)
		}
	}
	return s.wrapAllNodes(nodes...)
}

func (s *Selection) wrapAllNodes(ns ...*html.Node) *Selection {
	if len(ns) > 0 {
		return s.WrapAllNode(ns[0])
	}
	return s
}

func (s *Selection) WrapAllNode(n *html.Node) *Selection {
	if s.Size() == 0 {
		return s
	}

	wrap := cloneNode(n)

	first := s.Nodes[0]
	if first.Parent != nil {
		first.Parent.InsertBefore(wrap, first)
		first.Parent.RemoveChild(first)
	}

	for c := getFirstChildEl(wrap); c != nil; c = getFirstChildEl(wrap) {
		wrap = c
	}

	newSingleSelection(wrap, s.document).AppendSelection(s)

	return s
}

func (s *Selection) WrapInner(selector string) *Selection {
	return s.WrapInnerMatcher(compileMatcher(selector))
}

func (s *Selection) WrapInnerMatcher(m Matcher) *Selection {
	return s.wrapInnerNodes(m.MatchAll(s.document.rootNode)...)
}

func (s *Selection) WrapInnerSelection(sel *Selection) *Selection {
	return s.wrapInnerNodes(sel.Nodes...)
}

func (s *Selection) WrapInnerHtml(htmlStr string) *Selection {
	nodesMap := make(map[string][]*html.Node)
	for _, context := range s.Nodes {
		nodes, found := nodesMap[nodeName(context)]
		if !found {
			nodes = parseHtmlWithContext(htmlStr, context)
			nodesMap[nodeName(context)] = nodes
		}
		newSingleSelection(context, s.document).wrapInnerNodes(cloneNodes(nodes)...)
	}
	return s
}

func (s *Selection) WrapInnerNode(n *html.Node) *Selection {
	return s.wrapInnerNodes(n)
}

func (s *Selection) wrapInnerNodes(ns ...*html.Node) *Selection {
	if len(ns) == 0 {
		return s
	}

	s.Each(func(i int, s *Selection) {
		contents := s.Contents()

		if contents.Size() > 0 {
			contents.wrapAllNodes(ns...)
		} else {
			s.AppendNodes(cloneNode(ns[0]))
		}
	})

	return s
}

func parseHtml(h string) []*html.Node {
	nodes, err := html.ParseFragment(strings.NewReader(h), &html.Node{Type: html.ElementNode})
	if err != nil {
		panic("failed to parse HTML: " + err.Error())
	}
	return nodes
}

func parseHtmlWithContext(h string, context *html.Node) []*html.Node {
	nodes, err := html.ParseFragment(strings.NewReader(h), context)
	if err != nil {
		panic("failed to parse HTML: " + err.Error())
	}
	return nodes
}

func getFirstChildEl(n *html.Node) *html.Node {
	c := n.FirstChild
	for c != nil && c.Type != html.ElementNode {
		c = c.NextSibling
	}
	return c
}

func cloneNodes(ns []*html.Node) []*html.Node {
	cns := make([]*html.Node, 0, len(ns))

	for _, n := range ns {
		cns = append(cns, cloneNode(n))
	}

	return cns
}

func cloneNode(n *html.Node) *html.Node {
	nn := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     make([]html.Attribute, len(n.Attr)),
	}

	copy(nn.Attr, n.Attr)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nn.AppendChild(cloneNode(c))
	}

	return nn
}

func (s *Selection) manipulateNodes(ns []*html.Node, reverse bool,
	f func(sn *html.Node, n *html.Node)) *Selection {

	lasti := s.Size() - 1

	if reverse {
		for i, j := 0, len(ns)-1; i < j; i, j = i+1, j-1 {
			ns[i], ns[j] = ns[j], ns[i]
		}
	}

	for i, sn := range s.Nodes {
		for _, n := range ns {
			if i != lasti {
				f(sn, cloneNode(n))
			} else {
				if n.Parent != nil {
					n.Parent.RemoveChild(n)
				}
				f(sn, n)
			}
		}
	}

	return s
}

func (s *Selection) eachNodeHtml(htmlStr string, isParent bool, mergeFn func(n *html.Node, nodes []*html.Node)) *Selection {
	nodeCache := make(map[string][]*html.Node)
	var context *html.Node
	for _, n := range s.Nodes {
		if isParent {
			context = n.Parent
		} else {
			if n.Type != html.ElementNode {
				continue
			}
			context = n
		}
		if context != nil {
			nodes, found := nodeCache[nodeName(context)]
			if !found {
				nodes = parseHtmlWithContext(htmlStr, context)
				nodeCache[nodeName(context)] = nodes
			}
			mergeFn(n, cloneNodes(nodes))
		}
	}
	return s
}
