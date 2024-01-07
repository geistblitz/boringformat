package launder

import (
	"golang.org/x/net/html"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
	ToEnd   = maxInt
)

func (s *Selection) First() *Selection {
	return s.Eq(0)
}

func (s *Selection) Last() *Selection {
	return s.Eq(-1)
}

func (s *Selection) Eq(index int) *Selection {
	if index < 0 {
		index += len(s.Nodes)
	}

	if index >= len(s.Nodes) || index < 0 {
		return newEmptySelection(s.document)
	}

	return s.Slice(index, index+1)
}

func (s *Selection) Slice(start, end int) *Selection {
	if start < 0 {
		start += len(s.Nodes)
	}
	if end == ToEnd {
		end = len(s.Nodes)
	} else if end < 0 {
		end += len(s.Nodes)
	}
	return pushStack(s, s.Nodes[start:end])
}

func (s *Selection) Get(index int) *html.Node {
	if index < 0 {
		index += len(s.Nodes)
	}
	return s.Nodes[index]
}

func (s *Selection) Index() int {
	if len(s.Nodes) > 0 {
		return newSingleSelection(s.Nodes[0], s.document).PrevAll().Length()
	}
	return -1
}

func (s *Selection) IndexSelector(selector string) int {
	if len(s.Nodes) > 0 {
		sel := s.document.Find(selector)
		return indexInSlice(sel.Nodes, s.Nodes[0])
	}
	return -1
}

func (s *Selection) IndexMatcher(m Matcher) int {
	if len(s.Nodes) > 0 {
		sel := s.document.FindMatcher(m)
		return indexInSlice(sel.Nodes, s.Nodes[0])
	}
	return -1
}

func (s *Selection) IndexOfNode(node *html.Node) int {
	return indexInSlice(s.Nodes, node)
}

func (s *Selection) IndexOfSelection(sel *Selection) int {
	if sel != nil && len(sel.Nodes) > 0 {
		return indexInSlice(s.Nodes, sel.Nodes[0])
	}
	return -1
}
