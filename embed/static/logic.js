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

function consolePromptSubmit(event) {
  let p = document.getElementById("console-prompt");

  if (p.value === "") {
    return;
  }

  fetch(apiPath(CURRENT_NODE_PATH), {
    method: "POST",
    body: JSON.stringify({
      name: p.value,
    }),
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }
    })
    .then(() => {
      p.value = "";
    })
    .catch((error) => console.log(error));
}

function consoleSetDataFields(fields) {
  let console = document.getElementById("console-data-fields");
  console.innerHTML = "";

  fields.forEach((field) => {
    let div = document.createElement("div");
    div.classList.add("console-data-field");

    let span = document.createElement("span");
    span.className = "markdown-code";
    span.innerHTML = field;
    span.onclick = () => {
      navigator.clipboard.writeText(span.innerHTML);

      let prevBorder = span.style.border;
      span.style.border = "2px solid #ffffff";

      setTimeout(() => {
        span.style.border = prevBorder;
      }, 70);
    };

    div.appendChild(span);
    console.appendChild(div);
  });
}

function consoleSetNavigator(path) {
  let nav = document.getElementById("console-navigator");
  nav.innerHTML = "";

  let parts = path.split("/");

  for (let i = 0; i < parts.length; i++) {
    let span = document.createElement("span");
    span.className = "markdown-code console-navigator-field";
    span.innerHTML = parts[i];
    span.style.marginRight = "5px";
    span.onclick = () => {
      navigator.clipboard.writeText(span.innerHTML);

      let prevBorder = span.style.border;
      span.style.border = "2px solid #ffffff";

      setTimeout(() => {
        span.style.border = prevBorder;
      }, 70);

      setNode(parts.slice(0, i + 1).join("/"));
    };

    nav.appendChild(span);
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

function contentSet(content) {
  let c = document.getElementById("page-content");
  c.innerHTML = content;
}

// -------------------------------------------------------------------------- //
// leafs
// -------------------------------------------------------------------------- //

function setLeafs(leafPaths) {
  document.getElementById("leafs-container-one").innerHTML = "";
  document.getElementById("leafs-container-two").innerHTML = "";

  let getContainer = () => {
    let c1 = document.getElementById("leafs-container-one");
    let c2 = document.getElementById("leafs-container-two");

    if (c1.offsetHeight > c2.offsetHeight) {
      return c2;
    }

    return c1;
  };

  let newLeaf = (data) => {
    let leafBox = document.createElement("div");
    leafBox.className = "page-box clickable";
    leafBox.onclick = () => {
      setNode(data.path);
    };

    let leaf = document.createElement("div");
    leaf.className = "page-content-box";
    leaf.innerHTML = data.content;

    leafBox.appendChild(leaf);

    return leafBox;
  };

  let leafs = {};
  let all = [];

  leafPaths.forEach((path) => {
    all.push(
      getNode(path, "html", false).then((data) => {
        leafs[path] = newLeaf(data);
      })
    );
  });
  Promise.all(all).then(() => {
    for (let i = 0; i < leafPaths.length; i++) {
      getContainer().appendChild(leafs[leafPaths[i]]);
    }
  });
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

function setNode(path) {
  getNode(path, "html", true)
    .then((data) => {
      contentSet(data.content);
      consoleSetDataFields([data.id, data.name, data.path]);
      consoleSetNavigator(data.path);
      setLeafs(data.leafs);

      CURRENT_NODE_PATH = data.path;
      console.log(CURRENT_NODE_PATH);
      document.title = data.name;
    })
    .then(() => {
      console.log("loaded");
    });
}

setNode(CURRENT_NODE_PATH);
