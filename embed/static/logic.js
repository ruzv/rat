// -------------------------------------------------------------------------- //
// utils
// -------------------------------------------------------------------------- //

function apiPath(path) {
  if (path === "") {
    return window.location.origin + "/nodes/";
  }

  return window.location.origin + "/nodes/" + path + "/";
}

// -------------------------------------------------------------------------- //
// console
// -------------------------------------------------------------------------- //

function consoleControlsGetInput() {
  return document.getElementById("console-controls-input");
}

function consoleControlsAdd() {
  let input = consoleControlsGetInput();
  let v = input.value;
  input.value = "";

  if (v === "") {
    return;
  }

  fetch(apiPath(PATH), {
    method: "POST",
    body: JSON.stringify({
      name: v,
    }),
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then((response) => {
      // indicates whether the response is successful (status code 200-299) or not
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      return response.json();
    })
    .then((data) => {
      window.location = "/edit/" + data.path;
    })
    .catch((error) => console.log(error));
}

let deleteConfirmed = false;

function consoleControlsGetDelete() {
  return document.getElementById("console-controls-delete");
}

function consoleControlsDelete() {
  if (deleteConfirmed === false) {
    let d = consoleControlsGetDelete();
    d.innerHTML = "confirm";
    deleteConfirmed = true;

    return;
  }

  if (deleteConfirmed === true) {
    deleteConfirmed = false;
    fetch(apiPath(), {
      method: "DELETE",
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Request failed with status ${response.status}`);
        }
      })
      .catch((error) => console.log(error));
  }
}

class ConsoleData {
  constructor(id, name, path) {
    this.getIDField().innerHTML = id;
    this.getNameField().innerHTML = name;
    this.getPathField().innerHTML = path;
  }

  getIDField() {
    return document.getElementById("console-data-field-id");
  }

  getNameField() {
    return document.getElementById("console-data-field-name");
  }

  getPathField() {
    return document.getElementById("console-data-field-path");
  }
}

function copyConsoleDataField(element) {
  navigator.clipboard.writeText(element.innerHTML);

  let prevBorder = element.style.border;
  element.style.border = "2px solid #ffffff";

  setTimeout(() => {
    element.style.border = prevBorder;
  }, 70);
}

// -------------------------------------------------------------------------- //
// content
// -------------------------------------------------------------------------- //

class Content {
  constructor(html) {
    this.getElement().innerHTML = html;
  }

  getElement() {
    return document.getElementById("page-content");
  }
}

// -------------------------------------------------------------------------- //
// leafs
// -------------------------------------------------------------------------- //

class Leafs {
  constructor(leafPaths) {
    this.leafPaths = leafPaths;
    this.containerSwitcher = false;

    this.loadLeafs();
  }

  getContainer() {
    // id="leafs-container-two"
    let c;
    if (this.containerSwitcher == false) {
      c = document.getElementById("leafs-container-one");
    } else {
      c = document.getElementById("leafs-container-two");
    }

    this.containerSwitcher = !this.containerSwitcher;
    return c;
  }

  newLeaf(path) {
    // let leafLink = document.createElement("a");
    // leafLink.href = "/view/?node=" + path;

    let leafBox = document.createElement("div");
    leafBox.className = "page-box clickable";
    leafBox.onclick = () => {
      window.location = "/view/?node=" + path;
    };

    let leaf = document.createElement("div");
    leaf.className = "page-content-box";

    leafBox.appendChild(leaf);
    this.getContainer().appendChild(leafBox);

    return leaf;
  }

  loadLeafs() {
    this.leafPaths.forEach((path) => {
      let leaf = this.newLeaf(path);

      getNode(path, "html", "false", (data) => {
        leaf.innerHTML = data.content;
      });
    });
  }
}

function getNode(path, format, includeLeafs, callback) {
  let url = new URL(apiPath(path));

  url.searchParams.append("format", format);
  url.searchParams.append("leafs", includeLeafs);

  fetch(url, {
    method: "GET",
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      return response.json();
    })
    .then(callback)
    .catch((error) => console.log(error));
}

let content;
let consoleData;
let leafs;

getNode(PATH, "html", true, (data) => {
  content = new Content(data.content);
  consoleData = new ConsoleData(data.id, data.name, data.path);
  leafs = new Leafs(data.leafs);

  document.title = data.name;
});
