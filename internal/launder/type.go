package launder

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/geistblitz/boringformat/internal/launder/parser"
	"golang.org/x/net/html"
)

type Document struct {
	*Selection
	Url      *url.URL
	rootNode *html.Node
}

func NewDocumentFromNode(root *html.Node) *Document {
	return newDocument(root, nil)
}

func NewDocument(url string) (*Document, error) {
	res, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	return NewDocumentFromResponse(res)
}

func NewDocumentFromReader(r io.Reader) (*Document, error) {
	root, e := html.Parse(r)
	if e != nil {
		return nil, e
	}
	return newDocument(root, nil), nil
}

func NewDocumentFromResponse(res *http.Response) (*Document, error) {
	if res == nil {
		return nil, errors.New("Response is nil")
	}
	defer res.Body.Close()
	if res.Request == nil {
		return nil, errors.New("Response.Request is nil")
	}

	root, e := html.Parse(res.Body)
	if e != nil {
		return nil, e
	}

	return newDocument(root, res.Request.URL), nil
}

func CloneDocument(doc *Document) *Document {
	return newDocument(cloneNode(doc.rootNode), doc.Url)
}

func newDocument(root *html.Node, url *url.URL) *Document {
	d := &Document{nil, url, root}
	d.Selection = newSingleSelection(root, d)
	return d
}

type Selection struct {
	Nodes    []*html.Node
	document *Document
	prevSel  *Selection
}

func newEmptySelection(doc *Document) *Selection {
	return &Selection{nil, doc, nil}
}

func newSingleSelection(node *html.Node, doc *Document) *Selection {
	return &Selection{[]*html.Node{node}, doc, nil}
}

type Matcher interface {
	Match(*html.Node) bool
	MatchAll(*html.Node) []*html.Node
	Filter([]*html.Node) []*html.Node
}

func Single(selector string) Matcher {
	return singleMatcher{compileMatcher(selector)}
}

func SingleMatcher(m Matcher) Matcher {
	if _, ok := m.(singleMatcher); ok {
		return m
	}
	return singleMatcher{m}
}

func compileMatcher(s string) Matcher {
	cs, err := parser.Compile(s)
	if err != nil {
		return invalidMatcher{}
	}
	return cs
}

type singleMatcher struct {
	Matcher
}

func (m singleMatcher) MatchAll(n *html.Node) []*html.Node {
	if mm, ok := m.Matcher.(interface{ MatchFirst(*html.Node) *html.Node }); ok {
		node := mm.MatchFirst(n)
		if node == nil {
			return nil
		}
		return []*html.Node{node}
	}

	nodes := m.Matcher.MatchAll(n)
	if len(nodes) > 0 {
		return nodes[:1:1]
	}
	return nil
}

type invalidMatcher struct{}

func (invalidMatcher) Match(n *html.Node) bool             { return false }
func (invalidMatcher) MatchAll(n *html.Node) []*html.Node  { return nil }
func (invalidMatcher) Filter(ns []*html.Node) []*html.Node { return nil }
