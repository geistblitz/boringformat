package parser

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type Matcher interface {
	Match(n *html.Node) bool
}

type Sel interface {
	Matcher
	Specificity() Specificity

	String() string

	PseudoElement() string
}

func Parse(sel string) (Sel, error) {
	p := &parser{s: sel}
	compiled, err := p.parseSelector()
	if err != nil {
		return nil, err
	}

	if p.i < len(sel) {
		return nil, fmt.Errorf("parsing %q: %d bytes left over", sel, len(sel)-p.i)
	}

	return compiled, nil
}

func ParseWithPseudoElement(sel string) (Sel, error) {
	p := &parser{s: sel, acceptPseudoElements: true}
	compiled, err := p.parseSelector()
	if err != nil {
		return nil, err
	}

	if p.i < len(sel) {
		return nil, fmt.Errorf("parsing %q: %d bytes left over", sel, len(sel)-p.i)
	}

	return compiled, nil
}

func ParseGroup(sel string) (SelectorGroup, error) {
	p := &parser{s: sel}
	compiled, err := p.parseSelectorGroup()
	if err != nil {
		return nil, err
	}

	if p.i < len(sel) {
		return nil, fmt.Errorf("parsing %q: %d bytes left over", sel, len(sel)-p.i)
	}

	return compiled, nil
}

func ParseGroupWithPseudoElements(sel string) (SelectorGroup, error) {
	p := &parser{s: sel, acceptPseudoElements: true}
	compiled, err := p.parseSelectorGroup()
	if err != nil {
		return nil, err
	}

	if p.i < len(sel) {
		return nil, fmt.Errorf("parsing %q: %d bytes left over", sel, len(sel)-p.i)
	}

	return compiled, nil
}

type Selector func(*html.Node) bool

func Compile(sel string) (Selector, error) {
	compiled, err := ParseGroup(sel)
	if err != nil {
		return nil, err
	}

	return Selector(compiled.Match), nil
}

func MustCompile(sel string) Selector {
	compiled, err := Compile(sel)
	if err != nil {
		panic(err)
	}
	return compiled
}

func (s Selector) MatchAll(n *html.Node) []*html.Node {
	return s.matchAllInto(n, nil)
}

func (s Selector) matchAllInto(n *html.Node, storage []*html.Node) []*html.Node {
	if s(n) {
		storage = append(storage, n)
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		storage = s.matchAllInto(child, storage)
	}

	return storage
}

func queryInto(n *html.Node, m Matcher, storage []*html.Node) []*html.Node {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if m.Match(child) {
			storage = append(storage, child)
		}
		storage = queryInto(child, m, storage)
	}

	return storage
}

func QueryAll(n *html.Node, m Matcher) []*html.Node {
	return queryInto(n, m, nil)
}

func (s Selector) Match(n *html.Node) bool {
	return s(n)
}

func (s Selector) MatchFirst(n *html.Node) *html.Node {
	if s.Match(n) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		m := s.MatchFirst(c)
		if m != nil {
			return m
		}
	}
	return nil
}

func Query(n *html.Node, m Matcher) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if m.Match(c) {
			return c
		}
		if matched := Query(c, m); matched != nil {
			return matched
		}
	}

	return nil
}

func (s Selector) Filter(nodes []*html.Node) (result []*html.Node) {
	for _, n := range nodes {
		if s(n) {
			result = append(result, n)
		}
	}
	return result
}

func Filter(nodes []*html.Node, m Matcher) (result []*html.Node) {
	for _, n := range nodes {
		if m.Match(n) {
			result = append(result, n)
		}
	}
	return result
}

type tagSelector struct {
	tag string
}

func (t tagSelector) Match(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == t.tag
}

func (c tagSelector) Specificity() Specificity {
	return Specificity{0, 0, 1}
}

func (c tagSelector) PseudoElement() string {
	return ""
}

type classSelector struct {
	class string
}

func (t classSelector) Match(n *html.Node) bool {
	return matchAttribute(n, "class", func(s string) bool {
		return matchInclude(t.class, s, false)
	})
}

func (c classSelector) Specificity() Specificity {
	return Specificity{0, 1, 0}
}

func (c classSelector) PseudoElement() string {
	return ""
}

type idSelector struct {
	id string
}

func (t idSelector) Match(n *html.Node) bool {
	return matchAttribute(n, "id", func(s string) bool {
		return s == t.id
	})
}

func (c idSelector) Specificity() Specificity {
	return Specificity{1, 0, 0}
}

func (c idSelector) PseudoElement() string {
	return ""
}

type attrSelector struct {
	key, val, operation string
	regexp              *regexp.Regexp
	insensitive         bool
}

func (t attrSelector) Match(n *html.Node) bool {
	switch t.operation {
	case "":
		return matchAttribute(n, t.key, func(string) bool { return true })
	case "=":
		return matchAttribute(n, t.key, func(s string) bool { return matchInsensitiveValue(s, t.val, t.insensitive) })
	case "!=":
		return attributeNotEqualMatch(t.key, t.val, n, t.insensitive)
	case "~=":
		return matchAttribute(n, t.key, func(s string) bool { return matchInclude(t.val, s, t.insensitive) })
	case "|=":
		return attributeDashMatch(t.key, t.val, n, t.insensitive)
	case "^=":
		return attributePrefixMatch(t.key, t.val, n, t.insensitive)
	case "$=":
		return attributeSuffixMatch(t.key, t.val, n, t.insensitive)
	case "*=":
		return attributeSubstringMatch(t.key, t.val, n, t.insensitive)
	case "#=":
		return attributeRegexMatch(t.key, t.regexp, n)
	default:
		panic(fmt.Sprintf("unsuported operation : %s", t.operation))
	}
}

func matchInsensitiveValue(userAttr string, realAttr string, ignoreCase bool) bool {
	if ignoreCase {
		return strings.EqualFold(userAttr, realAttr)
	}
	return userAttr == realAttr

}

func matchAttribute(n *html.Node, key string, f func(string) bool) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == key && f(a.Val) {
			return true
		}
	}
	return false
}

func attributeNotEqualMatch(key, val string, n *html.Node, ignoreCase bool) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == key && matchInsensitiveValue(a.Val, val, ignoreCase) {
			return false
		}
	}
	return true
}

func matchInclude(val string, s string, ignoreCase bool) bool {
	for s != "" {
		i := strings.IndexAny(s, " \t\r\n\f")
		if i == -1 {
			return matchInsensitiveValue(s, val, ignoreCase)
		}
		if matchInsensitiveValue(s[:i], val, ignoreCase) {
			return true
		}
		s = s[i+1:]
	}
	return false
}

func attributeDashMatch(key, val string, n *html.Node, ignoreCase bool) bool {
	return matchAttribute(n, key,
		func(s string) bool {
			if matchInsensitiveValue(s, val, ignoreCase) {
				return true
			}
			if len(s) <= len(val) {
				return false
			}
			if matchInsensitiveValue(s[:len(val)], val, ignoreCase) && s[len(val)] == '-' {
				return true
			}
			return false
		})
}

func attributePrefixMatch(key, val string, n *html.Node, ignoreCase bool) bool {
	return matchAttribute(n, key,
		func(s string) bool {
			if strings.TrimSpace(s) == "" {
				return false
			}
			if ignoreCase {
				return strings.HasPrefix(strings.ToLower(s), strings.ToLower(val))
			}
			return strings.HasPrefix(s, val)
		})
}

func attributeSuffixMatch(key, val string, n *html.Node, ignoreCase bool) bool {
	return matchAttribute(n, key,
		func(s string) bool {
			if strings.TrimSpace(s) == "" {
				return false
			}
			if ignoreCase {
				return strings.HasSuffix(strings.ToLower(s), strings.ToLower(val))
			}
			return strings.HasSuffix(s, val)
		})
}

func attributeSubstringMatch(key, val string, n *html.Node, ignoreCase bool) bool {
	return matchAttribute(n, key,
		func(s string) bool {
			if strings.TrimSpace(s) == "" {
				return false
			}
			if ignoreCase {
				return strings.Contains(strings.ToLower(s), strings.ToLower(val))
			}
			return strings.Contains(s, val)
		})
}

func attributeRegexMatch(key string, rx *regexp.Regexp, n *html.Node) bool {
	return matchAttribute(n, key,
		func(s string) bool {
			return rx.MatchString(s)
		})
}

func (c attrSelector) Specificity() Specificity {
	return Specificity{0, 1, 0}
}

func (c attrSelector) PseudoElement() string {
	return ""
}

type neverMatchSelector struct {
	value string
}

func (s neverMatchSelector) Match(n *html.Node) bool {
	return false
}

func (s neverMatchSelector) Specificity() Specificity {
	return Specificity{0, 0, 0}
}

func (c neverMatchSelector) PseudoElement() string {
	return ""
}

type compoundSelector struct {
	selectors     []Sel
	pseudoElement string
}

func (t compoundSelector) Match(n *html.Node) bool {
	if len(t.selectors) == 0 {
		return n.Type == html.ElementNode
	}

	for _, sel := range t.selectors {
		if !sel.Match(n) {
			return false
		}
	}
	return true
}

func (s compoundSelector) Specificity() Specificity {
	var out Specificity
	for _, sel := range s.selectors {
		out = out.Add(sel.Specificity())
	}
	if s.pseudoElement != "" {
		out = out.Add(Specificity{0, 0, 1})
	}
	return out
}

func (c compoundSelector) PseudoElement() string {
	return c.pseudoElement
}

type combinedSelector struct {
	first      Sel
	combinator byte
	second     Sel
}

func (t combinedSelector) Match(n *html.Node) bool {
	if t.first == nil {
		return false
	}
	switch t.combinator {
	case 0:
		return t.first.Match(n)
	case ' ':
		return descendantMatch(t.first, t.second, n)
	case '>':
		return childMatch(t.first, t.second, n)
	case '+':
		return siblingMatch(t.first, t.second, true, n)
	case '~':
		return siblingMatch(t.first, t.second, false, n)
	default:
		panic("unknown combinator")
	}
}

func descendantMatch(a, d Matcher, n *html.Node) bool {
	if !d.Match(n) {
		return false
	}

	for p := n.Parent; p != nil; p = p.Parent {
		if a.Match(p) {
			return true
		}
	}

	return false
}

func childMatch(a, d Matcher, n *html.Node) bool {
	return d.Match(n) && n.Parent != nil && a.Match(n.Parent)
}

func siblingMatch(s1, s2 Matcher, adjacent bool, n *html.Node) bool {
	if !s2.Match(n) {
		return false
	}

	if adjacent {
		for n = n.PrevSibling; n != nil; n = n.PrevSibling {
			if n.Type == html.TextNode || n.Type == html.CommentNode {
				continue
			}
			return s1.Match(n)
		}
		return false
	}

	for c := n.PrevSibling; c != nil; c = c.PrevSibling {
		if s1.Match(c) {
			return true
		}
	}

	return false
}

func (s combinedSelector) Specificity() Specificity {
	spec := s.first.Specificity()
	if s.second != nil {
		spec = spec.Add(s.second.Specificity())
	}
	return spec
}

func (c combinedSelector) PseudoElement() string {
	if c.second == nil {
		return ""
	}
	return c.second.PseudoElement()
}

type SelectorGroup []Sel

func (s SelectorGroup) Match(n *html.Node) bool {
	for _, sel := range s {
		if sel.Match(n) {
			return true
		}
	}
	return false
}
