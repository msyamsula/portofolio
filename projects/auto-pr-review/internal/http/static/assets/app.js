const form = document.querySelector("#link-form");
const input = document.querySelector("#github-url");
const message = document.querySelector("#message");
const submitButton = document.querySelector("#submit-button");
const refreshButton = document.querySelector("#refresh-button");
const linksList = document.querySelector("#links-list");
const emptyState = document.querySelector("#empty-state");

function setMessage(text, kind = "") {
  message.textContent = text;
  message.className = kind ? `message ${kind}` : "message";
}

function setBusy(isBusy) {
  submitButton.disabled = isBusy;
  refreshButton.disabled = isBusy;
  for (const button of linksList.querySelectorAll(".delete-button")) {
    button.disabled = isBusy;
  }
}

function renderLinks(links) {
  linksList.replaceChildren();
  emptyState.hidden = links.length > 0;

  for (const link of links) {
    const item = document.createElement("li");
    item.className = "link-item";

    const id = document.createElement("span");
    id.className = "link-id";
    id.textContent = link.id;

    const anchor = document.createElement("a");
    anchor.className = "link-url";
    anchor.href = link.url;
    anchor.target = "_blank";
    anchor.rel = "noreferrer";
    anchor.textContent = link.url;

    const deleteButton = document.createElement("button");
    deleteButton.type = "button";
    deleteButton.className = "delete-button";
    deleteButton.dataset.id = link.id;
    deleteButton.textContent = "Delete";

    item.append(id, anchor, deleteButton);
    linksList.append(item);
  }
}

async function loadLinks() {
  const response = await fetch("/post", {
    headers: { Accept: "application/json" },
  });

  if (!response.ok) {
    throw new Error("Unable to load links");
  }

  const data = await response.json();
  renderLinks(data.links ?? []);
}

async function registerLink(event) {
  event.preventDefault();
  setBusy(true);
  setMessage("Saving link...");

  try {
    const response = await fetch("/post", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: input.value }),
    });
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error ?? "Unable to save link");
    }

    input.value = "";
    setMessage("Link saved.", "success");
    await loadLinks();
  } catch (error) {
    setMessage(error.message, "error");
  } finally {
    setBusy(false);
    input.focus();
  }
}

async function deleteLink(id) {
  setBusy(true);
  setMessage("Deleting link...");

  try {
    const response = await fetch(`/post/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: { Accept: "application/json" },
    });

    if (!response.ok) {
      let errorMessage = "Unable to delete link";
      try {
        const data = await response.json();
        errorMessage = data.error ?? errorMessage;
      } catch {
        // Keep the default error when the response has no JSON body.
      }
      throw new Error(errorMessage);
    }

    setMessage("Link deleted.", "success");
    await loadLinks();
  } catch (error) {
    setMessage(error.message, "error");
  } finally {
    setBusy(false);
  }
}

form.addEventListener("submit", registerLink);
linksList.addEventListener("click", (event) => {
  const button = event.target.closest(".delete-button");
  if (!button) {
    return;
  }

  deleteLink(button.dataset.id);
});
refreshButton.addEventListener("click", async () => {
  setBusy(true);
  setMessage("Refreshing links...");
  try {
    await loadLinks();
    setMessage("Links refreshed.", "success");
  } catch (error) {
    setMessage(error.message, "error");
  } finally {
    setBusy(false);
  }
});

loadLinks().catch((error) => setMessage(error.message, "error"));
