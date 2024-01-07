package launder

import "golang.org/x/net/html"

func (s *Selection) Add(selector string) *Selection {
	return s.AddNodes(findWithMatcher([]*html.Node{s.document.rootNode}, compileMatcher(selector))...)
}

func (s *Selection) AddMatcher(m Matcher) *Selection {
	return s.AddNodes(findWithMatcher([]*html.Node{s.document.rootNode}, m)...)
}

func (s *Selection) AddSelection(sel *Selection) *Selection {
	if sel == nil {
		return s.AddNodes()
	}
	return s.AddNodes(sel.Nodes...)
}

func (s *Selection) Union(sel *Selection) *Selection {
	return s.AddSelection(sel)
}

func (s *Selection) AddNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, appendWithoutDuplicates(s.Nodes, nodes, nil))
}

func (s *Selection) AndSelf() *Selection {
	return s.AddBack()
}

func (s *Selection) AddBack() *Selection {
	return s.AddSelection(s.prevSel)
}

func (s *Selection) AddBackFiltered(selector string) *Selection {
	return s.AddSelection(s.prevSel.Filter(selector))
}

func (s *Selection) AddBackMatcher(m Matcher) *Selection {
	return s.AddSelection(s.prevSel.FilterMatcher(m))
}
