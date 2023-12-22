---
id: e235bfd8-b55d-461d-8efb-1c301f02493e
---

# config

<rat graph depth=1 />

---

```yaml
port: 8888

# debug 0
# info  1
# warn  2
# error 3
logLevel: 0

port: 8888
logLevel: 0
services:
  provider:
    dir: /path/to/graph/dir
    enablePathCache: true
    root:
      content: hello
      template:
        name: "{{ .RawName }}"
        weight: 0
        content: |
          # {{ .Name }}

          <rat graph depth=1 />

          ---
  urlResolver:
    fileservers:
      - authority: http://fileserver:8080
        user: username
        password: passwd
  sync:
    repoDir: /path/to/dir/containing/.git
    interval: 10s
    keyPath: /path/to/key
    keyPassword: passwd
```
