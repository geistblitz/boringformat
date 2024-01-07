package launder

import "golang.org/x/net/html"

func (s *Selection) Filter(selector string) *Selection {
	return s.FilterMatcher(compileMatcher(selector))
}

func (s *Selection) FilterMatcher(m Matcher) *Selection {
	return pushStack(s, winnow(s, m, true))
}

func (s *Selection) Not(selector string) *Selection {
	return s.NotMatcher(compileMatcher(selector))
}

func (s *Selection) NotMatcher(m Matcher) *Selection {
	return pushStack(s, winnow(s, m, false))
}

func (s *Selection) FilterFunction(f func(int, *Selection) bool) *Selection {
	return pushStack(s, winnowFunction(s, f, true))
}

func (s *Selection) NotFunction(f func(int, *Selection) bool) *Selection {
	return pushStack(s, winnowFunction(s, f, false))
}

func (s *Selection) FilterNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, winnowNodes(s, nodes, true))
}

func (s *Selection) NotNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, winnowNodes(s, nodes, false))
}

func (s *Selection) FilterSelection(sel *Selection) *Selection {
	if sel == nil {
		return pushStack(s, winnowNodes(s, nil, true))
	}
	return pushStack(s, winnowNodes(s, sel.Nodes, true))
}

func (s *Selection) NotSelection(sel *Selection) *Selection {
	if sel == nil {
		return pushStack(s, winnowNodes(s, nil, false))
	}
	return pushStack(s, winnowNodes(s, sel.Nodes, false))
}

func (s *Selection) Intersection(sel *Selection) *Selection {
	return s.FilterSelection(sel)
}

func (s *Selection) Has(selector string) *Selection {
	return s.HasSelection(s.document.Find(selector))
}

func (s *Selection) HasMatcher(m Matcher) *Selection {
	return s.HasSelection(s.document.FindMatcher(m))
}

func (s *Selection) HasNodes(nodes ...*html.Node) *Selection {
	return s.FilterFunction(func(_ int, sel *Selection) bool {
		for _, n := range nodes {
			if sel.Contains(n) {
				return true
			}
		}
		return false
	})
}

func (s *Selection) HasSelection(sel *Selection) *Selection {
	if sel == nil {
		return s.HasNodes()
	}
	return s.HasNodes(sel.Nodes...)
}

func (s *Selection) End() *Selection {
	if s.prevSel != nil {
		return s.prevSel
	}
	return newEmptySelection(s.document)
}

func winnow(sel *Selection, m Matcher, keep bool) []*html.Node {
	if keep {
		return m.Filter(sel.Nodes)
	}
	return grep(sel, func(i int, s *Selection) bool {
		return !m.Match(s.Get(0))
	})
}

func winnowNodes(sel *Selection, nodes []*html.Node, keep bool) []*html.Node {
	if len(nodes)+len(sel.Nodes) < minNodesForSet {
		return grep(sel, func(i int, s *Selection) bool {
			return isInSlice(nodes, s.Get(0)) == keep
		})
	}

	set := make(map[*html.Node]bool)
	for _, n := range nodes {
		set[n] = true
	}
	return grep(sel, func(i int, s *Selection) bool {
		return set[s.Get(0)] == keep
	})
}

func winnowFunction(sel *Selection, f func(int, *Selection) bool, keep bool) []*html.Node {
	return grep(sel, func(i int, s *Selection) bool {
		return f(i, s) == keep
	})
}
