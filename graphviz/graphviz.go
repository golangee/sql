package graphviz

import "strings"

type Style struct{
	Style string
	Shape string
	FillColor string
	Color string
}

type Node struct{
	Id string
	Style *Style
	Label string
	Underline bool
}

type Graph struct{
	Nodes []*Node
	Edges map[string]string
}

func (g *Graph) String()string{
	sb := &strings.Builder{}
	sb.WriteString("digraph G {\n")
	sb.WriteString("//nodes\n")
	for _,node := range g.Nodes {
		sb.WriteString(node.Id)
		sb.WriteString(" [")
		sb.WriteString("]\n")
	}

	sb.WriteString("//edges\n")
	for from, to := range g.Edges {
		sb.WriteString(from+" -> "+to+";\n")
	}
	sb.WriteString("}\n")
}
