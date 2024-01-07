package launder

import "golang.org/x/net/html"

func (s *Selection) Is(selector string) bool {
	return s.IsMatcher(compileMatcher(selector))
}

func (s *Selection) IsMatcher(m Matcher) bool {
	if len(s.Nodes) > 0 {
		if len(s.Nodes) == 1 {
			return m.Match(s.Nodes[0])
		}
		return len(m.Filter(s.Nodes)) > 0
	}

	return false
}

func (s *Selection) IsFunction(f func(int, *Selection) bool) bool {
	return s.FilterFunction(f).Length() > 0
}

func (s *Selection) IsSelection(sel *Selection) bool {
	return s.FilterSelection(sel).Length() > 0
}

func (s *Selection) IsNodes(nodes ...*html.Node) bool {
	return s.FilterNodes(nodes...).Length() > 0
}

func (s *Selection) Contains(n *html.Node) bool {
	return sliceContains(s.Nodes, n)
}
