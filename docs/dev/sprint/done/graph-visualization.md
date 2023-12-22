---
id: a542fc08-4429-4971-93f0-0ef46e653ac5
---

# graph-visualization

<rat  graph />

could use https://www.graphviz.org/ .

generate a graph description in a simple format. done in
https://github.com/ruzv/rat/pull/7

use different visual generation comands to generate variations of graph

```sh
dot -Tpng test.graph -o output.png
neato -Tpng test.graph -o output.png
fdp -Tpng test.graph -o output.png
sfdp -Tpng test.graph -o output.png
```

blog explaining graph syntax
https://ncona.com/2020/06/create-diagrams-with-code-using-graphviz/

IDEA: graph visualization inside markdown. code block like `todo`, but
`graphviz`, that would contain a graphs definition, and server would render it
and display it.
