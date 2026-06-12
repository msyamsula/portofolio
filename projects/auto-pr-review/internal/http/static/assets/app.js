const form = document.querySelector("#link-form");
const input = document.querySelector("#github-url");
const message = document.querySelector("#message");
const submitButton = document.querySelector("#submit-button");
const refreshButton = document.querySelector("#refresh-button");
const linksList = document.querySelector("#links-list");
const emptyState = document.querySelector("#empty-state");

let jobPeriods = [];
let periodicJobs = new Map();

function setMessage(text, kind = "") {
  message.textContent = text;
  message.className = kind ? `message ${kind}` : "message";
}

function setBusy(isBusy) {
  submitButton.disabled = isBusy;
  refreshButton.disabled = isBusy;
  for (const control of linksList.querySelectorAll("button, select")) {
    control.disabled = isBusy;
  }
}

function setPeriodicJobs(jobs) {
  periodicJobs = new Map();
  for (const job of jobs) {
    periodicJobs.set(job.github_link, job.period);
  }
}

function renderLinks(links) {
  linksList.replaceChildren();
  emptyState.hidden = links.length > 0;

  for (const link of links) {
    const activePeriod = link.periodic_interval ?? periodicJobs.get(link.url) ?? "";
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

    const controls = document.createElement("div");
    controls.className = "row-controls";

    const periodSelect = document.createElement("select");
    periodSelect.className = activePeriod ? "period-select registered" : "period-select";
    periodSelect.dataset.url = link.url;
    periodSelect.dataset.activePeriod = activePeriod;

    const emptyOption = document.createElement("option");
    emptyOption.value = "";
    emptyOption.textContent = "No schedule";
    periodSelect.append(emptyOption);

    for (const period of jobPeriods) {
      const option = document.createElement("option");
      option.value = period;
      option.textContent = period;
      periodSelect.append(option);
    }
    periodSelect.value = activePeriod;

    const registeredButton = document.createElement("button");
    registeredButton.type = "button";
    registeredButton.className = "registered-button";
    registeredButton.dataset.url = link.url;
    registeredButton.dataset.period = activePeriod;
    registeredButton.hidden = !activePeriod;
    registeredButton.textContent = "Registered";

    const deleteButton = document.createElement("button");
    deleteButton.type = "button";
    deleteButton.className = "delete-button";
    deleteButton.dataset.id = link.id;
    deleteButton.textContent = "Delete";

    controls.append(periodSelect, registeredButton, deleteButton);
    item.append(id, anchor, controls);
    linksList.append(item);
  }
}

async function fetchJSON(url) {
  const response = await fetch(url, {
    headers: { Accept: "application/json" },
  });

  if (!response.ok) {
    throw new Error(`Unable to load ${url}`);
  }

  return response.json();
}

async function loadLinks() {
  const [periodData, jobsData, linksData] = await Promise.all([
    fetchJSON("/job-periods"),
    fetchJSON("/jobs"),
    fetchJSON("/post"),
  ]);

  jobPeriods = periodData.periods ?? [];
  setPeriodicJobs(jobsData.jobs ?? []);
  renderLinks(linksData.links ?? []);
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

async function registerPeriodicJob(url, period) {
  setBusy(true);
  setMessage("Registering periodic job...");

  try {
    const response = await fetch("/jobs", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url, period }),
    });
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error ?? "Unable to register periodic job");
    }

    setMessage("Periodic job registered.", "success");
    await loadLinks();
  } catch (error) {
    setMessage(error.message, "error");
  } finally {
    setBusy(false);
  }
}

async function deregisterPeriodicJob(url, period) {
  setBusy(true);
  setMessage("Deregistering periodic job...");

  try {
    const response = await fetch("/jobs", {
      method: "DELETE",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url, period }),
    });

    if (!response.ok) {
      let errorMessage = "Unable to deregister periodic job";
      try {
        const data = await response.json();
        errorMessage = data.error ?? errorMessage;
      } catch {
        // Keep the default error when the response has no JSON body.
      }
      throw new Error(errorMessage);
    }

    setMessage("Periodic job deregistered.", "success");
    await loadLinks();
  } catch (error) {
    setMessage(error.message, "error");
  } finally {
    setBusy(false);
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
linksList.addEventListener("change", (event) => {
  const select = event.target.closest(".period-select");
  if (!select) {
    return;
  }

  const activePeriod = select.dataset.activePeriod;
  if (select.value) {
    registerPeriodicJob(select.dataset.url, select.value);
    return;
  }
  if (activePeriod) {
    deregisterPeriodicJob(select.dataset.url, activePeriod);
  }
});
linksList.addEventListener("click", (event) => {
  const registeredButton = event.target.closest(".registered-button");
  if (registeredButton) {
    deregisterPeriodicJob(registeredButton.dataset.url, registeredButton.dataset.period);
    return;
  }

  const deleteButton = event.target.closest(".delete-button");
  if (!deleteButton) {
    return;
  }

  deleteLink(deleteButton.dataset.id);
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
