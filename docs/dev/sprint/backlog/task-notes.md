---
id: e12fb761-ae97-4d9e-8d39-d454ad472a28
---

# task-notes

<rat graph />

---

a graph node (note) that describes a task.

- create task nodes anywhere in the graph
- define task lists and/or task boards anywhere in the graph
- link task nodes to task lists or graphs
- modifying task node from list or graph would modifythe underlying linked task
  node
- allow linking a single task to different graphs or lists
- disallow duplicate tasks in graph or list

- task node
  - regular markdown node
  - header metadata
    - `type: task`
      - introduce `type: note` for regular notes
    - `compleated: true | false`
      - button in FE to toggle
    - `priority`
    - `created`
    - `compleated` - time
- todo integration
  - task nodes with todos or todo tokens are only considered complete when all
    todos are done
- task collectors
  - easy way to add tasks to collector. some kind of view of availible tasks and
    drag and drop
  - todo token like collector that searche sub nodes for tasks
  - task list
    - a todo like structure of task nodes
  - task board
    - kanban board of task nodes
    - changing columns changes task nodes parent.
    - compleated column / button
  - task tree
    - tasks and sub tasks
    - todo is like a sub acts like sub tasks
    - from graph structure?
