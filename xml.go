package main

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Using tree builder from
// https://stackoverflow.com/questions/30256729/how-to-traverse-through-xml-data-in-golang

// XMLNode describes a node in XML tree
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []*XMLNode `xml:",any"`
}

// UnmarshalXML recursively decodes XML into a tree
func (n *XMLNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node XMLNode
	return d.DecodeElement((*node)(n), &start)
}

func (n *XMLNode) childByName(childLocalName string) *XMLNode {
	for _, n = range n.Nodes {
		if n.XMLName.Local == childLocalName {
			return n
		}
	}
	return nil
}

func (n *XMLNode) String() string {
	s := serAttributes(n.Attrs)
	start := "<"
	end := "/>"
	if len(n.Nodes) > 0 {
		end = ">"
	}
	if len(s) > 0 {
		return start + n.XMLName.Local + " " + s + end
	}
	return start + n.XMLName.Local + s + end
}

func hasChildNode(n *XMLNode, childLocalName string) bool {
	if n.childByName(childLocalName) == nil {
		return false
	}
	return true
}

func serAttributes(attrs []xml.Attr) string {
	if len(attrs) == 0 {
		return ""
	}
	var a []string
	for _, attr := range attrs {
		s := fmt.Sprintf("%s=%s", attr.Name.Local, esc(attr.Value))
		a = append(a, s)
	}
	return strings.Join(a, " ")
}

func esc(s string) string {
	if strings.Contains(s, " ") {
		return fmt.Sprintf(`"%s"`, s)
	}
	return s
}

func xmlSerializeNodeRecur(node *XMLNode, level int) []string {
	indent := strings.Repeat("  ", level)
	if len(node.Nodes) == 0 {
		return []string{indent + node.String()}
	}
	lines := []string{indent + node.String()}
	for _, n := range node.Nodes {
		childLines := xmlSerializeNodeRecur(n, level+1)
		lines = append(lines, childLines...)
	}
	s := fmt.Sprintf("%s</%s>", indent, node.XMLName.Local)
	return append(lines, s)
}

func xmlSerializeNode(node *XMLNode, level int) string {
	lines := xmlSerializeNodeRecur(node, level)
	return strings.Join(lines, "\n")
}

func xmlPrintFunc(node *XMLNode, level int) bool {
	fmt.Printf("%s%s\n", strings.Repeat("  ", level), node)
	return true
}

func xmlPrint(root *XMLNode) {
	nodes := []*XMLNode{root}
	xmlWalk(nodes, 0, xmlPrintFunc)
}

func xmlWalk(nodes []*XMLNode, level int, f func(*XMLNode, int) bool) {
	for _, n := range nodes {
		if f(n, level) {
			xmlWalk(n.Nodes, level+1, f)
		}
	}
}

func mustNoAttrs(attrs []xml.Attr, n *XMLNode) {
	if n != nil {
		//panicIf(len(attrs) != 0, "unsupported attribute in node '%s'", xmlSerializeNode(n, 0))
		panicIf(len(attrs) != 0, "unsupported attribute in node '%s'", n)
	} else {
		panicIf(len(attrs) != 0, "unsupported attribute: '%s'", serAttributes(attrs))
	}
}

func mustNoChildren(n *XMLNode) {
	panicIf(len(n.Nodes) > 0, "unexpected children in node:\n%s\n", xmlSerializeNode(n, 0))
}

func mustTag(is string, wanted string) {
	if is != wanted {
		err := fmt.Errorf("tag is '%s', expected: '%s'", is, wanted)
		must(err)
	}
}

// Attrs combines attributes and their owner
// Helper for extracting attributes
type Attrs struct {
	Attrs []xml.Attr
	Owner *XMLNode
}

// NewAttrs creates Attrs from owner
func NewAttrs(owner *XMLNode) *Attrs {
	return &Attrs{
		Attrs: owner.Attrs,
		Owner: owner,
	}
}

func (a *Attrs) mustEmpty() {
	if len(a.Attrs) > 0 {
		s := a.Attrs[0].Name
		panicIf(true, "Unsupported attribute %s in node:\n%s", s, a.Owner)
	}
}

func (a *Attrs) extractByName(attrName string) *xml.Attr {
	var rest []xml.Attr
	var res *xml.Attr
	for i, attr := range a.Attrs {
		if attr.Name.Local == attrName {
			res = &a.Attrs[i]
		} else {
			rest = append(rest, attr)
		}
	}
	a.Attrs = rest
	return res
}

func (a *Attrs) extractByNameI(attrName string) *xml.Attr {
	var rest []xml.Attr
	var res *xml.Attr
	for i, attr := range a.Attrs {
		if strings.EqualFold(attr.Name.Local, attrName) {
			res = &a.Attrs[i]
		} else {
			rest = append(rest, attr)
		}
	}
	a.Attrs = rest
	return res
}

func (a *Attrs) mustExtractString(attrName string) string {
	attr := a.extractByName(attrName)
	panicIf(attr == nil, "missing attribute '%s' in node:\n%s", attrName, a.Owner)
	return attr.Value
}

func (a *Attrs) mustExtractInt(attrName string) int {
	s := a.mustExtractString(attrName)
	i, err := strconv.Atoi(s)
	must(err)
	return i
}

func (a *Attrs) extractString(name string) string {
	attr := a.extractByName(name)
	if attr == nil {
		return ""
	}
	return attr.Value
}

func (a *Attrs) extractInt(name string, defValue int) int {
	attr := a.extractByName(name)
	if attr == nil {
		return defValue
	}
	n, err := strconv.Atoi(attr.Value)
	must(err)
	return n
}

func (a *Attrs) extractBool(name string) bool {
	attr := a.extractByName(name)
	if attr == nil {
		return false
	}
	return mustParseBool(attr.Value)
}
