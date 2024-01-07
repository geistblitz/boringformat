package parser

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type abstractPseudoClass struct{}

func (s abstractPseudoClass) Specificity() Specificity {
	return Specificity{0, 1, 0}
}

func (c abstractPseudoClass) PseudoElement() string {
	return ""
}

type relativePseudoClassSelector struct {
	name  string
	match SelectorGroup
}

func (s relativePseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch s.name {
	case "not":
		return !s.match.Match(n)
	case "has":
		return hasDescendantMatch(n, s.match)
	case "haschild":
		return hasChildMatch(n, s.match)
	default:
		panic(fmt.Sprintf("unsupported relative pseudo class selector : %s", s.name))
	}
}

func hasChildMatch(n *html.Node, a Matcher) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if a.Match(c) {
			return true
		}
	}
	return false
}

func hasDescendantMatch(n *html.Node, a Matcher) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if a.Match(c) || (c.Type == html.ElementNode && hasDescendantMatch(c, a)) {
			return true
		}
	}
	return false
}

func (s relativePseudoClassSelector) Specificity() Specificity {
	var max Specificity
	for _, sel := range s.match {
		newSpe := sel.Specificity()
		if max.Less(newSpe) {
			max = newSpe
		}
	}
	return max
}

func (c relativePseudoClassSelector) PseudoElement() string {
	return ""
}

type containsPseudoClassSelector struct {
	abstractPseudoClass
	value string
	own   bool
}

func (s containsPseudoClassSelector) Match(n *html.Node) bool {
	var text string
	if s.own {
		text = strings.ToLower(nodeOwnText(n))
	} else {
		text = strings.ToLower(nodeText(n))
	}
	return strings.Contains(text, s.value)
}

type regexpPseudoClassSelector struct {
	abstractPseudoClass
	regexp *regexp.Regexp
	own    bool
}

func (s regexpPseudoClassSelector) Match(n *html.Node) bool {
	var text string
	if s.own {
		text = nodeOwnText(n)
	} else {
		text = nodeText(n)
	}
	return s.regexp.MatchString(text)
}

func writeNodeText(n *html.Node, b *bytes.Buffer) {
	switch n.Type {
	case html.TextNode:
		b.WriteString(n.Data)
	case html.ElementNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNodeText(c, b)
		}
	}
}

func nodeText(n *html.Node) string {
	var b bytes.Buffer
	writeNodeText(n, &b)
	return b.String()
}

func nodeOwnText(n *html.Node) string {
	var b bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

type nthPseudoClassSelector struct {
	abstractPseudoClass
	a, b         int
	last, ofType bool
}

func (s nthPseudoClassSelector) Match(n *html.Node) bool {
	if s.a == 0 {
		if s.last {
			return simpleNthLastChildMatch(s.b, s.ofType, n)
		} else {
			return simpleNthChildMatch(s.b, s.ofType, n)
		}
	}
	return nthChildMatch(s.a, s.b, s.last, s.ofType, n)
}

func nthChildMatch(a, b int, last, ofType bool, n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	parent := n.Parent
	if parent == nil {
		return false
	}

	i := -1
	count := 0
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		if (c.Type != html.ElementNode) || (ofType && c.Data != n.Data) {
			continue
		}
		count++
		if c == n {
			i = count
			if !last {
				break
			}
		}
	}

	if i == -1 {
		return false
	}

	if last {
		i = count - i + 1
	}

	i -= b
	if a == 0 {
		return i == 0
	}

	return i%a == 0 && i/a >= 0
}

func simpleNthChildMatch(b int, ofType bool, n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	parent := n.Parent
	if parent == nil {
		return false
	}

	count := 0
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || (ofType && c.Data != n.Data) {
			continue
		}
		count++
		if c == n {
			return count == b
		}
		if count >= b {
			return false
		}
	}
	return false
}

func simpleNthLastChildMatch(b int, ofType bool, n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	parent := n.Parent
	if parent == nil {
		return false
	}

	count := 0
	for c := parent.LastChild; c != nil; c = c.PrevSibling {
		if c.Type != html.ElementNode || (ofType && c.Data != n.Data) {
			continue
		}
		count++
		if c == n {
			return count == b
		}
		if count >= b {
			return false
		}
	}
	return false
}

type onlyChildPseudoClassSelector struct {
	abstractPseudoClass
	ofType bool
}

func (s onlyChildPseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	parent := n.Parent
	if parent == nil {
		return false
	}

	count := 0
	for c := parent.FirstChild; c != nil; c = c.NextSibling {
		if (c.Type != html.ElementNode) || (s.ofType && c.Data != n.Data) {
			continue
		}
		count++
		if count > 1 {
			return false
		}
	}

	return count == 1
}

type inputPseudoClassSelector struct {
	abstractPseudoClass
}

func (s inputPseudoClassSelector) Match(n *html.Node) bool {
	return n.Type == html.ElementNode && (n.Data == "input" || n.Data == "select" || n.Data == "textarea" || n.Data == "button")
}

type emptyElementPseudoClassSelector struct {
	abstractPseudoClass
}

func (s emptyElementPseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.ElementNode:
			return false
		case html.TextNode:
			if strings.TrimSpace(nodeText(c)) == "" {
				continue
			} else {
				return false
			}
		}
	}

	return true
}

type rootPseudoClassSelector struct {
	abstractPseudoClass
}

func (s rootPseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if n.Parent == nil {
		return false
	}
	return n.Parent.Type == html.DocumentNode
}

func hasAttr(n *html.Node, attr string) bool {
	return matchAttribute(n, attr, func(string) bool { return true })
}

type linkPseudoClassSelector struct {
	abstractPseudoClass
}

func (s linkPseudoClassSelector) Match(n *html.Node) bool {
	return (n.DataAtom == atom.A || n.DataAtom == atom.Area || n.DataAtom == atom.Link) && hasAttr(n, "href")
}

type langPseudoClassSelector struct {
	abstractPseudoClass
	lang string
}

func (s langPseudoClassSelector) Match(n *html.Node) bool {
	own := matchAttribute(n, "lang", func(val string) bool {
		return val == s.lang || strings.HasPrefix(val, s.lang+"-")
	})
	if n.Parent == nil {
		return own
	}
	return own || s.Match(n.Parent)
}

type enabledPseudoClassSelector struct {
	abstractPseudoClass
}

func (s enabledPseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch n.DataAtom {
	case atom.A, atom.Area, atom.Link:
		return hasAttr(n, "href")
	case atom.Optgroup, atom.Menuitem, atom.Fieldset:
		return !hasAttr(n, "disabled")
	case atom.Button, atom.Input, atom.Select, atom.Textarea, atom.Option:
		return !hasAttr(n, "disabled") && !inDisabledFieldset(n)
	}
	return false
}

type disabledPseudoClassSelector struct {
	abstractPseudoClass
}

func (s disabledPseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch n.DataAtom {
	case atom.Optgroup, atom.Menuitem, atom.Fieldset:
		return hasAttr(n, "disabled")
	case atom.Button, atom.Input, atom.Select, atom.Textarea, atom.Option:
		return hasAttr(n, "disabled") || inDisabledFieldset(n)
	}
	return false
}

func hasLegendInPreviousSiblings(n *html.Node) bool {
	for s := n.PrevSibling; s != nil; s = s.PrevSibling {
		if s.DataAtom == atom.Legend {
			return true
		}
	}
	return false
}

func inDisabledFieldset(n *html.Node) bool {
	if n.Parent == nil {
		return false
	}
	if n.Parent.DataAtom == atom.Fieldset && hasAttr(n.Parent, "disabled") &&
		(n.DataAtom != atom.Legend || hasLegendInPreviousSiblings(n)) {
		return true
	}
	return inDisabledFieldset(n.Parent)
}

type checkedPseudoClassSelector struct {
	abstractPseudoClass
}

func (s checkedPseudoClassSelector) Match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch n.DataAtom {
	case atom.Input, atom.Menuitem:
		return hasAttr(n, "checked") && matchAttribute(n, "type", func(val string) bool {
			t := toLowerASCII(val)
			return t == "checkbox" || t == "radio"
		})
	case atom.Option:
		return hasAttr(n, "selected")
	}
	return false
}
