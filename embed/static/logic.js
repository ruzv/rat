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
    // this.leafPaths = leafPaths;
    this.containerSwitcher = false;

    this.leftColumnHeight = 0;
    this.rightColumnHeight = 0;

    this.loadLeafs(leafPaths);
  }

  getContainer() {
    let c1 = document.getElementById("leafs-container-one");
    let c2 = document.getElementById("leafs-container-two");

    if (c1.offsetHeight > c2.offsetHeight) {
      return c2;
    }

    // console.log(
    //   "get",
    //   document.getElementById("leafs-container-one").offsetHeight,
    //   document.getElementById("leafs-container-two").offsetHeight
    // );

    return c1;
  }

  newLeaf(data) {
    // let leafLink = document.createElement("a");
    // leafLink.href = "/view/?node=" + path;

    let leafBox = document.createElement("div");
    leafBox.className = "page-box clickable";
    leafBox.onclick = () => {
      window.location = "/view/?node=" + data.path;
    };

    let leaf = document.createElement("div");
    leaf.className = "page-content-box";
    leaf.innerHTML = data.content;

    leafBox.appendChild(leaf);

    return leafBox;
  }

  loadLeafs(leafPaths) {
    let leafs = {};

    let all = [];

    for (let i = 0; i < leafPaths.length; i++) {
      all.push(
        getNode(leafPaths[i], "html", false).then((data) => {
          leafs[i] = this.newLeaf(data);
        })
      );
    }
    Promise.all(all).then(() => {
      for (let i = 0; i < leafPaths.length; i++) {
        this.getContainer().appendChild(leafs[i]);
      }
    });
  }
}

function getNode(path, format, includeLeafs) {
  let url = new URL(apiPath(path));

  url.searchParams.append("format", format);
  url.searchParams.append("leafs", includeLeafs);

  return fetch(url, {
    method: "GET",
  }).then((response) => {
    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`);
    }

    return response.json();
  });
}

let content;
let consoleData;
let leafs;

getNode(PATH, "html", true)
  .then((data) => {
    content = new Content(data.content);
    consoleData = new ConsoleData(data.id, data.name, data.path);

    if (data.leafs) {
      leafs = new Leafs(data.leafs);
    }

    document.title = data.name;
  })
  .then(() => {
    console.log("loaded");
  });
