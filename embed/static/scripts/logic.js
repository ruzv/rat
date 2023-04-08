let Title, NodeID, NodeName, NodePath;

const Prompt = (
  elementID,
  callbacks = {
    submit: () => {},
    close: () => {},
    navigateUp: () => {},
    navigateDown: () => {},
    keydown: () => {},
  }
) => {
  let p = {
    element: document.getElementById(elementID),
    mode: "insert",
    cursorPos: 0,

    callbacks: callbacks,
  };

  p.element.addEventListener("keydown", (event) => {
    if (event.key === "Enter") {
      p.callbacks.submit();
      return;
    }

    if (p.mode === "normal") {
      let pos;

      switch (event.key) {
        case "i":
          p.mode = "insert";
          break;
        case "y": // left
          pos = Math.max(0, p.element.selectionStart - 1);
          p.element.selectionStart = pos;
          p.element.selectionEnd = pos;
          break;
        case "o": // right
          pos = Math.min(p.element.selectionStart + 1, p.element.value.length);
          p.element.selectionStart = pos;
          p.element.selectionEnd = pos;
          break;
        case "e": // up
          p.callbacks.navigateUp();
          break;
        case "n": // down
          p.callbacks.navigateDown();
          break;
        case "Escape":
          p.callbacks.close();
          break;
      }

      event.preventDefault();
      return;
    }

    if (p.mode === "insert") {
      if (event.key === "Escape") {
        p.mode = "normal";
        return;
      }

      p.callbacks.keydown();
      return;
    }
  });

  return p;
};

const Modal = (elementID) => {
  return {
    element: document.getElementById(elementID),

    show() {
      this.element.style.display = "block";
    },
    hide() {
      this.element.style.display = "none";
    },
  };
};

const SearchResults = (elementID) => {
  return {
    element: document.getElementById(elementID),
    selected: -1,

    submit() {
      if (this.selected === -1) {
        return;
      }

      navigateToPath(this.element.childNodes[this.selected]);
    },

    select(index) {
      if (this.element.childNodes.length === 0) {
        return;
      }

      if (index < 0) {
        index = 0;
      }

      if (index >= this.element.childNodes.length) {
        index = this.element.childNodes.length - 1;
      }

      if (this.selected !== -1) {
        this.element.childNodes[this.selected].className = "search-result";
      }

      this.element.childNodes[index].className = "search-result-selected";
      this.selected = index;
    },

    selectUp() {
      this.select(this.selected - 1);
    },

    selectDown() {
      this.select(this.selected + 1);
    },

    add(path) {
      let result = ((path) => {
        let result = document.createElement("div");
        result.innerHTML = path;
        result.className = "search-result";
        result.setAttribute("data-path", path);

        return result;
      })(path);

      this.element.appendChild(result);
    },
    clear() {
      this.element.innerHTML = "";
    },
  };
};

const Search = () => {
  let s = {
    modal: Modal("search-modal"),
    prompt: Prompt("search-prompt"),
    results: SearchResults("search-results"),
    selectedResult: 0,

    show() {
      this.modal.show();
      this.prompt.element.focus();
      this.prompt.element.select();
      this.prompt.mode = "insert";
    },
  };

  s.prompt.callbacks.submit = () => s.results.submit();
  s.prompt.callbacks.close = () => s.modal.hide();
  s.prompt.callbacks.navigateUp = () => s.results.selectUp();
  s.prompt.callbacks.navigateDown = () => s.results.selectDown();

  fetch("/graph/index/", { method: "GET" })
    .then((response) => {
      if (!response.ok) {
        response.json().then((data) => {
          throw new Error(
            `failed to get graph index, status ${response.status}, body ${data}`
          );
        });
      }

      return response.json();
    })
    .then((data) => {
      s.results.clear();

      data.forEach((path) => {
        s.results.add(path);
      });

      s.results.select(0);

      s.prompt.callbacks.keydown = () => {
        s.results.clear();

        data.forEach((path) => {
          if (!path.includes(s.prompt.element.value)) {
            return;
          }

          s.results.add(path);
        });

        s.results.select(0);
      };
    })
    .catch((error) => console.log(error));

  return s;
};

const NewNode = () => {
  let nn = {
    modal: Modal("new-node-modal"),
    prompt: Prompt("new-node-prompt"),

    show() {
      this.modal.show();
      this.prompt.element.focus();
      this.prompt.element.select();
      this.prompt.mode = "insert";
    },
  };

  nn.prompt.callbacks.close = () => {
    nn.modal.hide();
  };
  nn.prompt.callbacks.submit = () => {
    if (nn.prompt.element.value === "") {
      return;
    }

    fetch(
      new URL("/graph/nodes/" + NodePath + "/", window.location.origin).href,
      {
        method: "POST",
        body: JSON.stringify({
          name: nn.prompt.element.value,
        }),
        headers: {
          "Content-Type": "application/json",
        },
      }
    )
      .then((response) => {
        if (!response.ok) {
          response.json().then((data) => {
            throw new Error(
              `create new node request failed, ` +
                `status ${response.status}, body ${data}`
            );
          });
        }
      })
      .then(() => {
        window.location.reload();
      })
      .catch((error) => console.log(error));
  };

  return nn;
};

// -------------------------------------------------------------------------- //
// console
// -------------------------------------------------------------------------- //

function consolePromptSubmit() {
  let p = document.getElementById("console-prompt");

  if (p.value === "") {
    return;
  }

  fetch(
    new URL("/graph/nodes/" + NodePath + "/", window.location.origin).href,
    {
      method: "POST",
      body: JSON.stringify({
        name: p.value,
      }),
      headers: {
        "Content-Type": "application/json",
      },
    }
  )
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }
    })
    .then(() => {
      window.location.reload();
    })
    .catch((error) => console.log(error));
}

function consoleSetNavigator(path) {
  let nav = document.getElementById("console-navigator");
  nav.innerHTML = "";

  let parts = path.split("/");

  for (let i = 0; i < parts.length; i++) {
    let div = document.createElement("div");

    div.className = "console-field clickable";
    div.innerHTML = parts[i];
    div.style.marginRight = "5px";

    div.onclick = () => {
      let path = parts.slice(0, i + 1).join("/");

      window.location.href = new URL(
        "/view/" + path,
        window.location.origin
      ).href;
    };

    nav.appendChild(div);
  }
}

function navigateToPath(e) {
  window.location.href = new URL(
    "/view/" + e.getAttribute("data-path"),
    window.location.origin
  ).href;
}

function innerHTMLToClipboard(element) {
  navigator.clipboard.writeText(element.innerHTML.trim());
}

// a colletion of functions that are used to handle the kanban board
// drag and drop
const KanbanHandlers = {
  dragstart: (event) => {
    // Add the target element's id to the data transfer object
    event.dataTransfer.setData("text/plain", event.target.id);
  },
  dragover: (event) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  },
  drop: (event) => {
    event.preventDefault();

    let card = document.getElementById(
      event.dataTransfer.getData("text/plain")
    );

    let target = event.target;

    while (!target.className.includes("kanban-column")) {
      target = target.parentElement;
    }

    target.appendChild(card);

    fetch("/graph/move/" + card.getAttribute("data-id"), {
      method: "POST",
      body: JSON.stringify({
        new_path:
          target.getAttribute("data-path") +
          "/" +
          card.getAttribute("data-name"),
      }),
      headers: {
        "Content-Type": "application/json",
      },
    })
      .then((response) => {
        if (!response.ok) {
          response.json().then((data) => {
            throw new Error(
              `failed to move kanban card, ` +
                `status ${response.status}, body ${data}`
            );
          });
        }

        return response.json();
      })
      .catch((error) => console.log(error));
  },
};

document.addEventListener("DOMContentLoaded", function () {
  Title = document.getElementById("title");
  NodeID = Title.getAttribute("data-id");
  NodeName = Title.getAttribute("data-name");
  NodePath = Title.getAttribute("data-path");

  let search = Search();
  let newNode = NewNode();

  consoleSetNavigator(NodePath);

  shortcut.add("Meta+P", function () {
    search.show();
  });

  shortcut.add("Meta+Shift+P", function () {
    newNode.show();
  });
});
