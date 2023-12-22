---
id: 9d201712-28c2-41ea-90c5-33ffeeefe500
---

# templates

<rat graph />

---

new nodes can be templated.

when creating a new node a walk up the graph tree is performed searchin for a
parent node that has a template defined. root node always must have a template.

here's what can be templated

- `name` the name of a sub node. usefull for yournaling or daily notes, where
  you would want to include the date, time in the nodes name.
- `weight` - how should the sub nodes be weighted. also usefull in a daily notes
  type of situation, because sorting notes lexicographically can produce daily
  note orders like
  - day1
  - day10
  - day2
- `content` - what content should the sub node begin with.
- `template` - recursive template. define what template should the sub node
  have. usefull when classifying frequent nodes. the following structure defines
  a `month/day` node naming structure

```yml
template:
  weight: "{{ .Month }}"
  name: "{{ .Month }}"
  content: |
    <!-- cSpell:disable -->

    # {{ .Name }} `{{ .Smile }}`

    <rat graph />

    ---
  template:
    weight: "{{ .Day }}"
    name: "{{ .Day }}"
    content: |
      <!-- cSpell:disable -->

      # {{ .Name }}

      <rat graph />

      ---
```

the following template fields are available

- `Name` - the name of the newly created node
- `RawName` - the name you inputed when creating node
- `Day`
- `Month`
- `Year`
- `Week`
- `YearDay`
- `Smile`
