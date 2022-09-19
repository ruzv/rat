// -------------------------------------------------------------------------- //
// utils
// -------------------------------------------------------------------------- //

function apiPath() {
  return "/graphs/" + PATH;
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

  fetch(apiPath(), {
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
        throw new Error(`Request failed with status ${reponse.status}`);
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

  update() {
    fetch(apiPath(), {
      method: "GET",
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Request failed with status ${response.status}`);
        }

        return response.json();
      })
      .then((data) => {
        this.getElement().innerHTML = data.html;
      })
      .catch((error) => console.log(error));
  }
}

let content;
let consoleData;

function getNode() {
  fetch(apiPath(), {
    method: "GET",
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      return response.json();
    })
    .then((data) => {
      // "id":       n.ID.String(),
      // "name":     n.Name,
      // "path":     n.Path,
      // "raw":      n.Content,
      // "markdown": n.Markdown(),
      // "html":     n.HTML(),

      content = new Content(data.html);
      consoleData = new ConsoleData(data.id, data.name, data.path);
      // editor = new Editor(data.raw, content);
    })
    .catch((error) => console.log(error));
}

getNode();
