package launder

import (
	"bytes"
	"io"

	"golang.org/x/net/html"
)

const minNodesForSet = 1000

var nodeNames = []string{
	html.ErrorNode:    "#error",
	html.TextNode:     "#text",
	html.DocumentNode: "#document",
	html.CommentNode:  "#comment",
}

func NodeName(s *Selection) string {
	if s.Length() == 0 {
		return ""
	}
	return nodeName(s.Get(0))
}

func nodeName(node *html.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case html.ElementNode, html.DoctypeNode:
		return node.Data
	default:
		if int(node.Type) < len(nodeNames) {
			return nodeNames[node.Type]
		}
		return ""
	}
}

func Render(w io.Writer, s *Selection) error {
	if s.Length() == 0 {
		return nil
	}
	n := s.Get(0)
	return html.Render(w, n)
}

func OuterHtml(s *Selection) (string, error) {
	var buf bytes.Buffer
	if err := Render(&buf, s); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func sliceContains(container []*html.Node, contained *html.Node) bool {
	for _, n := range container {
		if nodeContains(n, contained) {
			return true
		}
	}

	return false
}

func nodeContains(container *html.Node, contained *html.Node) bool {
	for contained = contained.Parent; contained != nil; contained = contained.Parent {
		if container == contained {
			return true
		}
	}
	return false
}

func isInSlice(slice []*html.Node, node *html.Node) bool {
	return indexInSlice(slice, node) > -1
}

func indexInSlice(slice []*html.Node, node *html.Node) int {
	if node != nil {
		for i, n := range slice {
			if n == node {
				return i
			}
		}
	}
	return -1
}

func appendWithoutDuplicates(target []*html.Node, nodes []*html.Node, targetSet map[*html.Node]bool) []*html.Node {
	if targetSet == nil && len(target)+len(nodes) < minNodesForSet {
		for _, n := range nodes {
			if !isInSlice(target, n) {
				target = append(target, n)
			}
		}
		return target
	}

	if targetSet == nil {
		targetSet = make(map[*html.Node]bool, len(target))
		for _, n := range target {
			targetSet[n] = true
		}
	}
	for _, n := range nodes {
		if !targetSet[n] {
			target = append(target, n)
			targetSet[n] = true
		}
	}

	return target
}

func grep(sel *Selection, predicate func(i int, s *Selection) bool) (result []*html.Node) {
	for i, n := range sel.Nodes {
		if predicate(i, newSingleSelection(n, sel.document)) {
			result = append(result, n)
		}
	}
	return result
}

func pushStack(fromSel *Selection, nodes []*html.Node) *Selection {
	result := &Selection{nodes, fromSel.document, fromSel}
	return result
}
