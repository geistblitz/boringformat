package launder

import "golang.org/x/net/html"

type siblingType int

const (
	siblingPrevUntil siblingType = iota - 3
	siblingPrevAll
	siblingPrev
	siblingAll
	siblingNext
	siblingNextAll
	siblingNextUntil
	siblingAllIncludingNonElements
)

func (s *Selection) Find(selector string) *Selection {
	return pushStack(s, findWithMatcher(s.Nodes, compileMatcher(selector)))
}

func (s *Selection) FindMatcher(m Matcher) *Selection {
	return pushStack(s, findWithMatcher(s.Nodes, m))
}

func (s *Selection) FindSelection(sel *Selection) *Selection {
	if sel == nil {
		return pushStack(s, nil)
	}
	return s.FindNodes(sel.Nodes...)
}

func (s *Selection) FindNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, mapNodes(nodes, func(i int, n *html.Node) []*html.Node {
		if sliceContains(s.Nodes, n) {
			return []*html.Node{n}
		}
		return nil
	}))
}

func (s *Selection) Contents() *Selection {
	return pushStack(s, getChildrenNodes(s.Nodes, siblingAllIncludingNonElements))
}

func (s *Selection) ContentsFiltered(selector string) *Selection {
	if selector != "" {
		return s.ChildrenFiltered(selector)
	}
	return s.Contents()
}

func (s *Selection) ContentsMatcher(m Matcher) *Selection {
	return s.ChildrenMatcher(m)
}

func (s *Selection) Children() *Selection {
	return pushStack(s, getChildrenNodes(s.Nodes, siblingAll))
}

func (s *Selection) ChildrenFiltered(selector string) *Selection {
	return filterAndPush(s, getChildrenNodes(s.Nodes, siblingAll), compileMatcher(selector))
}

func (s *Selection) ChildrenMatcher(m Matcher) *Selection {
	return filterAndPush(s, getChildrenNodes(s.Nodes, siblingAll), m)
}

func (s *Selection) Parent() *Selection {
	return pushStack(s, getParentNodes(s.Nodes))
}

func (s *Selection) ParentFiltered(selector string) *Selection {
	return filterAndPush(s, getParentNodes(s.Nodes), compileMatcher(selector))
}

func (s *Selection) ParentMatcher(m Matcher) *Selection {
	return filterAndPush(s, getParentNodes(s.Nodes), m)
}

func (s *Selection) Closest(selector string) *Selection {
	cs := compileMatcher(selector)
	return s.ClosestMatcher(cs)
}

func (s *Selection) ClosestMatcher(m Matcher) *Selection {
	return pushStack(s, mapNodes(s.Nodes, func(i int, n *html.Node) []*html.Node {
		for ; n != nil; n = n.Parent {
			if m.Match(n) {
				return []*html.Node{n}
			}
		}
		return nil
	}))
}

func (s *Selection) ClosestNodes(nodes ...*html.Node) *Selection {
	set := make(map[*html.Node]bool)
	for _, n := range nodes {
		set[n] = true
	}
	return pushStack(s, mapNodes(s.Nodes, func(i int, n *html.Node) []*html.Node {
		for ; n != nil; n = n.Parent {
			if set[n] {
				return []*html.Node{n}
			}
		}
		return nil
	}))
}

func (s *Selection) ClosestSelection(sel *Selection) *Selection {
	if sel == nil {
		return pushStack(s, nil)
	}
	return s.ClosestNodes(sel.Nodes...)
}

func (s *Selection) Parents() *Selection {
	return pushStack(s, getParentsNodes(s.Nodes, nil, nil))
}

func (s *Selection) ParentsFiltered(selector string) *Selection {
	return filterAndPush(s, getParentsNodes(s.Nodes, nil, nil), compileMatcher(selector))
}

func (s *Selection) ParentsMatcher(m Matcher) *Selection {
	return filterAndPush(s, getParentsNodes(s.Nodes, nil, nil), m)
}

func (s *Selection) ParentsUntil(selector string) *Selection {
	return pushStack(s, getParentsNodes(s.Nodes, compileMatcher(selector), nil))
}

func (s *Selection) ParentsUntilMatcher(m Matcher) *Selection {
	return pushStack(s, getParentsNodes(s.Nodes, m, nil))
}

func (s *Selection) ParentsUntilSelection(sel *Selection) *Selection {
	if sel == nil {
		return s.Parents()
	}
	return s.ParentsUntilNodes(sel.Nodes...)
}

func (s *Selection) ParentsUntilNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, getParentsNodes(s.Nodes, nil, nodes))
}

func (s *Selection) ParentsFilteredUntil(filterSelector, untilSelector string) *Selection {
	return filterAndPush(s, getParentsNodes(s.Nodes, compileMatcher(untilSelector), nil), compileMatcher(filterSelector))
}

func (s *Selection) ParentsFilteredUntilMatcher(filter, until Matcher) *Selection {
	return filterAndPush(s, getParentsNodes(s.Nodes, until, nil), filter)
}

func (s *Selection) ParentsFilteredUntilSelection(filterSelector string, sel *Selection) *Selection {
	return s.ParentsMatcherUntilSelection(compileMatcher(filterSelector), sel)
}

func (s *Selection) ParentsMatcherUntilSelection(filter Matcher, sel *Selection) *Selection {
	if sel == nil {
		return s.ParentsMatcher(filter)
	}
	return s.ParentsMatcherUntilNodes(filter, sel.Nodes...)
}

func (s *Selection) ParentsFilteredUntilNodes(filterSelector string, nodes ...*html.Node) *Selection {
	return filterAndPush(s, getParentsNodes(s.Nodes, nil, nodes), compileMatcher(filterSelector))
}

func (s *Selection) ParentsMatcherUntilNodes(filter Matcher, nodes ...*html.Node) *Selection {
	return filterAndPush(s, getParentsNodes(s.Nodes, nil, nodes), filter)
}

func (s *Selection) Siblings() *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingAll, nil, nil))
}

func (s *Selection) SiblingsFiltered(selector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingAll, nil, nil), compileMatcher(selector))
}

func (s *Selection) SiblingsMatcher(m Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingAll, nil, nil), m)
}

func (s *Selection) Next() *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingNext, nil, nil))
}

func (s *Selection) NextFiltered(selector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNext, nil, nil), compileMatcher(selector))
}

func (s *Selection) NextMatcher(m Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNext, nil, nil), m)
}

func (s *Selection) NextAll() *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingNextAll, nil, nil))
}

func (s *Selection) NextAllFiltered(selector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNextAll, nil, nil), compileMatcher(selector))
}

func (s *Selection) NextAllMatcher(m Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNextAll, nil, nil), m)
}

func (s *Selection) Prev() *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingPrev, nil, nil))
}

func (s *Selection) PrevFiltered(selector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrev, nil, nil), compileMatcher(selector))
}

func (s *Selection) PrevMatcher(m Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrev, nil, nil), m)
}

func (s *Selection) PrevAll() *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingPrevAll, nil, nil))
}

func (s *Selection) PrevAllFiltered(selector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrevAll, nil, nil), compileMatcher(selector))
}

func (s *Selection) PrevAllMatcher(m Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrevAll, nil, nil), m)
}

func (s *Selection) NextUntil(selector string) *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		compileMatcher(selector), nil))
}

func (s *Selection) NextUntilMatcher(m Matcher) *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		m, nil))
}

func (s *Selection) NextUntilSelection(sel *Selection) *Selection {
	if sel == nil {
		return s.NextAll()
	}
	return s.NextUntilNodes(sel.Nodes...)
}

func (s *Selection) NextUntilNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		nil, nodes))
}

func (s *Selection) PrevUntil(selector string) *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		compileMatcher(selector), nil))
}

func (s *Selection) PrevUntilMatcher(m Matcher) *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		m, nil))
}

func (s *Selection) PrevUntilSelection(sel *Selection) *Selection {
	if sel == nil {
		return s.PrevAll()
	}
	return s.PrevUntilNodes(sel.Nodes...)
}

func (s *Selection) PrevUntilNodes(nodes ...*html.Node) *Selection {
	return pushStack(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		nil, nodes))
}

func (s *Selection) NextFilteredUntil(filterSelector, untilSelector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		compileMatcher(untilSelector), nil), compileMatcher(filterSelector))
}

func (s *Selection) NextFilteredUntilMatcher(filter, until Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		until, nil), filter)
}

func (s *Selection) NextFilteredUntilSelection(filterSelector string, sel *Selection) *Selection {
	return s.NextMatcherUntilSelection(compileMatcher(filterSelector), sel)
}

func (s *Selection) NextMatcherUntilSelection(filter Matcher, sel *Selection) *Selection {
	if sel == nil {
		return s.NextMatcher(filter)
	}
	return s.NextMatcherUntilNodes(filter, sel.Nodes...)
}

func (s *Selection) NextFilteredUntilNodes(filterSelector string, nodes ...*html.Node) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		nil, nodes), compileMatcher(filterSelector))
}

func (s *Selection) NextMatcherUntilNodes(filter Matcher, nodes ...*html.Node) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingNextUntil,
		nil, nodes), filter)
}

func (s *Selection) PrevFilteredUntil(filterSelector, untilSelector string) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		compileMatcher(untilSelector), nil), compileMatcher(filterSelector))
}

func (s *Selection) PrevFilteredUntilMatcher(filter, until Matcher) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		until, nil), filter)
}

func (s *Selection) PrevFilteredUntilSelection(filterSelector string, sel *Selection) *Selection {
	return s.PrevMatcherUntilSelection(compileMatcher(filterSelector), sel)
}

func (s *Selection) PrevMatcherUntilSelection(filter Matcher, sel *Selection) *Selection {
	if sel == nil {
		return s.PrevMatcher(filter)
	}
	return s.PrevMatcherUntilNodes(filter, sel.Nodes...)
}

func (s *Selection) PrevFilteredUntilNodes(filterSelector string, nodes ...*html.Node) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		nil, nodes), compileMatcher(filterSelector))
}

func (s *Selection) PrevMatcherUntilNodes(filter Matcher, nodes ...*html.Node) *Selection {
	return filterAndPush(s, getSiblingNodes(s.Nodes, siblingPrevUntil,
		nil, nodes), filter)
}

func filterAndPush(srcSel *Selection, nodes []*html.Node, m Matcher) *Selection {
	sel := &Selection{nodes, srcSel.document, nil}
	return pushStack(srcSel, winnow(sel, m, true))
}

func findWithMatcher(nodes []*html.Node, m Matcher) []*html.Node {
	return mapNodes(nodes, func(i int, n *html.Node) (result []*html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				result = append(result, m.MatchAll(c)...)
			}
		}
		return
	})
}

func getParentsNodes(nodes []*html.Node, stopm Matcher, stopNodes []*html.Node) []*html.Node {
	return mapNodes(nodes, func(i int, n *html.Node) (result []*html.Node) {
		for p := n.Parent; p != nil; p = p.Parent {
			sel := newSingleSelection(p, nil)
			if stopm != nil {
				if sel.IsMatcher(stopm) {
					break
				}
			} else if len(stopNodes) > 0 {
				if sel.IsNodes(stopNodes...) {
					break
				}
			}
			if p.Type == html.ElementNode {
				result = append(result, p)
			}
		}
		return
	})
}

func getSiblingNodes(nodes []*html.Node, st siblingType, untilm Matcher, untilNodes []*html.Node) []*html.Node {
	var f func(*html.Node) bool

	if st == siblingNextUntil || st == siblingPrevUntil {
		f = func(n *html.Node) bool {
			if untilm != nil {
				sel := newSingleSelection(n, nil)
				return sel.IsMatcher(untilm)
			} else if len(untilNodes) > 0 {
				sel := newSingleSelection(n, nil)
				return sel.IsNodes(untilNodes...)
			}
			return false
		}
	}

	return mapNodes(nodes, func(i int, n *html.Node) []*html.Node {
		return getChildrenWithSiblingType(n.Parent, st, n, f)
	})
}

func getChildrenNodes(nodes []*html.Node, st siblingType) []*html.Node {
	return mapNodes(nodes, func(i int, n *html.Node) []*html.Node {
		return getChildrenWithSiblingType(n, st, nil, nil)
	})
}

func getChildrenWithSiblingType(parent *html.Node, st siblingType, skipNode *html.Node,
	untilFunc func(*html.Node) bool) (result []*html.Node) {

	var iter = func(cur *html.Node) (ret *html.Node) {
		for {
			switch st {
			case siblingAll, siblingAllIncludingNonElements:
				if cur == nil {
					if ret = parent.FirstChild; ret == skipNode && skipNode != nil {
						ret = skipNode.NextSibling
					}
				} else {
					if ret = cur.NextSibling; ret == skipNode && skipNode != nil {
						ret = skipNode.NextSibling
					}
				}
			case siblingPrev, siblingPrevAll, siblingPrevUntil:
				if cur == nil {
					ret = skipNode.PrevSibling
				} else {
					ret = cur.PrevSibling
				}
			case siblingNext, siblingNextAll, siblingNextUntil:
				if cur == nil {
					ret = skipNode.NextSibling
				} else {
					ret = cur.NextSibling
				}
			default:
				panic("Invalid sibling type.")
			}
			if ret == nil || ret.Type == html.ElementNode || st == siblingAllIncludingNonElements {
				return
			}
			cur = ret
		}
	}

	for c := iter(nil); c != nil; c = iter(c) {
		if st == siblingNextUntil || st == siblingPrevUntil {
			if untilFunc(c) {
				return
			}
		}
		result = append(result, c)
		if st == siblingNext || st == siblingPrev {
			return
		}
	}
	return
}

func getParentNodes(nodes []*html.Node) []*html.Node {
	return mapNodes(nodes, func(i int, n *html.Node) []*html.Node {
		if n.Parent != nil && n.Parent.Type == html.ElementNode {
			return []*html.Node{n.Parent}
		}
		return nil
	})
}

func mapNodes(nodes []*html.Node, f func(int, *html.Node) []*html.Node) (result []*html.Node) {
	set := make(map[*html.Node]bool)
	for i, n := range nodes {
		if vals := f(i, n); len(vals) > 0 {
			result = appendWithoutDuplicates(result, vals, set)
		}
	}
	return result
}
