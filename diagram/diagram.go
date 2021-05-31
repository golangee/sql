// Copyright 2021 The Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diagram

import (
	"fmt"
	"github.com/golangee/sql/ddl"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"

	"github.com/emicklei/dot"
)

type LayoutAlgorithm string

const (
	Dot   LayoutAlgorithm = "dot"
	Neato LayoutAlgorithm = "neato"
	TwoPi LayoutAlgorithm = "twopi"
	Circo LayoutAlgorithm = "circo"
	Fdp   LayoutAlgorithm = "fdp"
)

// GenerateDot creates a graphviz representation of the relationships
// between the given tables.
func GenerateDot(tables []ddl.Table, layout LayoutAlgorithm) string {
	g := dot.NewGraph(dot.Undirected)
	g.Attr("layout", layout)
	g.Attr("overlap", "false") // can be "false", "scale", "true"

	// Collect table nodes while drawing here
	tableNodes := make(map[string]dot.Node)

	// Draw tables and their attributes
	for _, table := range tables {
		tableID := table.Name // A tables name is unique, so it can be used as an id
		tableNode := g.Node(tableID).Label(table.Name).Box()
		tableNodes[tableID] = tableNode

		for _, column := range table.Columns {
			name := fmt.Sprintf("%s: %s", column.Name, column.Type)
			columnNode := g.Node(randomNodeID())

			// Underline PRIMARY KEY by using Literal
			if column.PrimaryKey {
				columnNode.Attr("label", dot.Literal("<<u>"+name+"</u>>"))
			} else {
				columnNode.Label(name)
			}

			tableNode.Edge(columnNode)
		}
	}

	// Draw FOREIGN KEYs as diamonds connecting tables
	for _, table := range tables {
		for _, foreignKey := range table.ForeignKeys {
			relNode := g.Node(randomNodeID())
			relNode.Attr("shape", "diamond")

			if foreignKey.Name != nil {
				relNode.Label(*foreignKey.Name)
			} else {
				// This constraint does not have a name, use "ref" as a default
				relNode.Label("ref")
			}

			g.Edge(tableNodes[table.Name], relNode).Label("N")
			g.Edge(relNode, tableNodes[foreignKey.ReferenceTable]).Label("1")
		}
	}

	return g.String()
}

// DotToSvg converts dot representation of a graph to an svg image
// using the "dot" executable.
func DotToSvg(dot string) (string, error) {
	dotCommand := exec.Command("dot", "-Tsvg")
	dotCommand.Stdin = strings.NewReader(dot)

	output, err := dotCommand.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run dot command: %w", err)
	}

	return string(output), nil
}

// Nodes with the same name will be processed as the same node by graphviz.
// To prevent that, we can assign them random IDs.
func randomNodeID() string {
	r := rand.Int63()

	return strconv.FormatInt(r, 16)
}
