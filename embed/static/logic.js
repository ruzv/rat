let Title = document.getElementById("title");

let NodeID = Title.getAttribute("data-id");
let NodeName = Title.getAttribute("data-name");
let NodePath = Title.getAttribute("data-path");

// -------------------------------------------------------------------------- //
// console
// -------------------------------------------------------------------------- //

function consolePromptSubmit() {
  let p = document.getElementById("console-prompt");

  if (p.value === "") {
    return;
  }

  fetch(new URL("/nodes/" + NodePath + "/", window.location.origin).href, {
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

consoleSetNavigator(NodePath);
