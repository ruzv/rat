---
id: 2e996058-52fe-4120-9557-897ea06a717a
---

# bugs

```todo

- rat tag directly below a heading, gets wrongly pased
  example:
    # nhnhnn
    <rat graph />
    no new line

- allowed to create nodes with emtpy name

- in FE, when navigating with search from one note with a kanban, to another,
  the first notes cards are rendered in the second notes kanban.

- add version rat tag

- handle api http errors with httpuril.Error

- some api error responses get CORS blocked, register a blobal CORS policy
  also figure out nice way to handle OPTIONS requests.

- when starting rat with port already in use, seerver enters infiate loop
  implement too many restarts, too quickly logic.

- can't use emoji in node name

- creating a new node with a name that contains a `/` fails with a 500.
  should be a 400.


- code blocks with html are rendered in browser, not displayed

- wrong formatting when creating a new node. header has new line

- update rat markdown lib
  https://github.com/ruzv/rat/security/dependabot/1

- fix node links `[](uuid)`

x fails when moving node to a node that does not have any sub nodes
x codeblock white spaces dont get rendered, can be seen in
  [ ](500a20fd-7442-41c6-ad49-2fd5def928f2)
x pages without leafs, dont have title set
x on mac command+k and command+shift+k does not open the search and create
  modals
```

## code block rendering broken when codeblock in list

markdown lib parser problem
markdown@v0.0.0-20220731190611-dcdaee8e7a53/parser/block.go

example ->

- long life certs
  - update `$(step path)/config/ca.json` `.authority.claims` field
  - ```json
    {
      "claims": {
        "minTLSCertDuration": "5m",
        "maxTLSCertDuration": "1000h",
        "defaultTLSCertDuration": "1000h"
      }
    }
    ```
    ```json
        hello
    ```
  - request
    ```sh
    step ca certificate 192.168.64.5 srv.crt srv.key --not-after=1000h
    ```
