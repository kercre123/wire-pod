function checkLanguage() {
  fetch("/api/get_stt_info")
    .then((response) => response.json())
    .then((parsed) => {
      const sectionLanguage = document.getElementById("section-language");
      const languageSelection = document.getElementById("languageSelection");

      if (parsed.provider !== "vosk" && parsed.provider !== "whisper.cpp") {
        console.log("stt not vosk/whisper");
        sectionLanguage.style.display = "none";
        languageSelection.value = "en-US";
      } else {
        sectionLanguage.style.display = "block";
        console.log(parsed.language);
        languageSelection.value = "en-US";
      }
    });
}

function updateSetupStatus(statusString) {
  const setupStatus = document.getElementById("setup-status");
  setupStatus.innerHTML = `<p>${statusString}</p>`;
}

function sendSetupInfo() {
  document.getElementById("config-options").style.display = "none";
  updateSetupStatus("Initiating setup...");

  const language = document.getElementById("languageSelection").value;
  const langData = { language };

  document.getElementById("languageSelectionDiv").style.display = "none";

  fetch("/api/set_stt_info", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(langData),
  })
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        updateSetupStatus("Language set successfully.");
        initWeatherAPIKey();
      } else if (response.includes("downloading")) {
        updateSetupStatus("Downloading language model...");
        var interval = setInterval(() => {
          fetch("/api/get_download_status")
            .then((response) => response.text())
            .then((statusText) => {
              updateSetupStatus(statusText);
              if (statusText.includes("success")) {
                updateSetupStatus("Language set successfully.");
                initWeatherAPIKey();
                clearInterval(interval);
              } else if (statusText.includes("error")) {
                document.getElementById("config-options").style.display = "block";
                clearInterval(interval);
              } else if (statusText.includes("not downloading")) {
                updateSetupStatus("Initiating language model download...");
              }
            });
        }, 500);
      } else if (response.includes("vosk")) {
        initWeatherAPIKey();
      } else if (response.includes("error")) {
        updateSetupStatus(response);
        document.getElementById("config-options").style.display = "block";
      }
    });
}

function initWeatherAPIKey() {
  const provider = document.getElementById("weatherProvider").value;
  if (provider) {
    updateSetupStatus("Setting weather API key...");
    const apiKey = document.getElementById("weatherAPIAddForm").elements["apiKey"].value;
    const data = { provider, key: apiKey };

    fetch("/api/set_weather_api", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    })
      .then((response) => response.text())
      .then((response) => {
        updateSetupStatus(response);
        initKGAPIKey();
      });
  } else {
    initKGAPIKey();
  }
}

function initKGAPIKey() {
  const provider = getE("kgProvider").value;
  if (provider == "") {
    setConn();
    return
  }
  const key = getE(`${provider}Key`).value;
  let doEnable = true;
  let model = "";
  let openAIPrompt = "";
  let id = "";
  let intentgraph = getE("intentyes").checked
  let saveChat = getE("saveChatYes").checked
  let doCommands = getE("commandYes").checked
  let endpoint = "";

  if (!key) {
    alert("You must provide an API key.");
    return;
  }

  if (provider === "custom") {
    model = getE("customModel").value;
    openAIPrompt = getE("customAIPrompt").value;
    endpoint = getE("customAIEndpoint").value;

    if (!model) {
      alert("You must provide an LLM model.");
      return;
    }
    if (!endpoint) {
      alert("You must provide an LLM endpoint.");
      return;
    }
  } else if (provider === "together") {
    model = getE("togetherModel").value;
    openAIPrompt = getE("togetherAIPrompt").value;
  } else if (provider === "openai") {
    openAIPrompt = getE("openAIPrompt").value;
  } else if (provider === "houndify") {
    id = getE("houndID").value;
    if (!id) {
      alert("You must provide a client ID.");
      return;
    }
  } else {
    doEnable = false;
  }

  const data = {
    enable: doEnable,
    provider,
    key,
    model,
    id,
    intentgraph: intentgraph,
    robotName: "",
    openai_prompt: openAIPrompt,
    save_chat: saveChat,
    commands_enable: doCommands,
    endpoint,
  };

  fetch("/api/set_kg_api", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      updateSetupStatus(response);
      setConn();
    });
}

function checkConn() {
  const connValue = document.getElementById("connSelection").value;
  document.getElementById("portViz").style.display = connValue === "ip" ? "block" : "none";
}

function setConn() {
  updateSetupStatus("Setting connection type (ep or ip)...");
  const connValue = document.getElementById("connSelection").value;
  let port = document.getElementById("portInput").value;
  port = port ? port : "443";
  const url = connValue === "ep" ? "/api-chipper/use_ep" : `/api-chipper/use_ip?port=${port}`;

  fetch(url)
    .then((response) => response.text())
    .then((response) => {
      if (response) {
        updateSetupStatus("Setup is complete! Wire-pod has started. Redirecting to main page...");
        setTimeout(() => window.location.href = "/", 3000);
      } else {
        updateSetupStatus("Error setting up wire-pod, check the logs");
        document.getElementById("config-options").style.display = "block";
      }
    });
}

function directToIndex() {
  window.location.href = "/index.html";
}