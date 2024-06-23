function checkLanguage() {
  let xhr = new XMLHttpRequest();
  xhr.open("GET", "/api/get_stt_info");
  xhr.send();
  xhr.onload = function () {
    parsed = JSON.parse(xhr.response);
    if (
      parsed["sttProvider"] != "vosk" &&
      parsed["sttProvider"] != "whisper.cpp"
    ) {
      console.log("stt not vosk/whisper");
      document.getElementById("section-language").style.display = "none";
      document.getElementById("languageSelection").value = "en-US";
    } else {
      document.getElementById("section-language").style.display = "block";
      console.log(parsed["sttLanguage"]);
      document.getElementById("languageSelection").value = "en-US";
    }
  };
}

function updateSetupStatus(statusString) {
  setupStatus = document.getElementById("setup-status");
  setupStatus.innerHTML = "";
  setupStatusP = document.createElement("p");
  setupStatusP.innerHTML = statusString;
  setupStatus.appendChild(setupStatusP);
}

function sendSetupInfo() {
  document.getElementById("config-options").style.display = "none";
  updateSetupStatus("Initiating setup...");
  language = document.getElementById("languageSelection").value;

  // set language first
  var langData = "language=" + language;
  document.getElementById("languageSelectionDiv").style.display = "none";
  fetch("/api/set_stt_info?" + langData)
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        updateSetupStatus("Language set successfully.");
        initWeatherAPIKey();
      } else if (response.includes("downloading")) {
        updateSetupStatus("Downloading language model...");
        inte = setInterval(function () {
          fetch("/api/get_download_status")
            .then((response) => response.text())
            .then((response) => {
              statusText = response;
              if (response.includes("success")) {
                updateSetupStatus("Language set successfully.");
                initWeatherAPIKey();
                clearInterval(inte);
              } else if (response.includes("error")) {
                document.getElementById("config-options").style.display =
                  "block";
              } else if (response.includes("not downloading")) {
                statusText = "Initiating language model download...";
              }
              updateSetupStatus(statusText);
            });
        }, 500);
      } else if (response.includes("vosk")) {
        initWeatherAPIKey();
      } else if (response.includes("error")) {
        updateSetupStatus(response);
        document.getElementById("config-options").style.display = "block";
        return;
      }
    });
}

function initWeatherAPIKey() {
  var provider = document.getElementById("weatherProvider").value;
  if (provider != "") {
    updateSetupStatus("Setting weather API key...");
    var form = document.getElementById("weatherAPIAddForm");

    var data =
      "provider=" + provider + "&api_key=" + form.elements["apiKey"].value;
    fetch("/api/set_weather_api?" + data)
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
  var provider = getE("kgProvider").value;
  var key = "";
  var openAIPrompt = "";
  var id = "";
  var intentgraph = "";
  var robotName = "";
  var model = "";
  var saveChat = "";
  var doCommands = "";
  var endpoint = "";

  if (provider == "openai") {
    key = getE("openAIKey").value;
    if (key == "") {
      alert("You must provide an API key.")
      return
    }
    openAIPrompt = getE("openAIPrompt").value;
    if (getE("commandYes").checked == true) {
      doCommands = "true";
    } else {
      doCommands = "false";
    }
    if (getE("intentyes").checked == true) {
      intentgraph = "true";
    } else {
      intentgraph = "false";
    }
    if (getE("saveChatYes").checked == true) {
      saveChat = "true";
    } else {
      saveChat = "false";
    }
  } else if (provider == "custom") {
    key = getE("customAIKey").value;
    model = getE("customModel").value;
    if (model == "") {
      alert("You must provide an LLM model.")
    }
    openAIPrompt = getE("customAIPrompt").value;
    endpoint = getE("customAIEndpoint").value;
    if (key == "") {
      alert("You must provide an API key.")
      return
    }
    if (endpoint == "") {
      alert("You must provide an LLM endpoint.")
      return
    }
    if (getE("commandYes").checked == true) {
      doCommands = "true";
    } else {
      doCommands = "false";
    }
    if (getE("intentyes").checked == true) {
      intentgraph = "true";
    } else {
      intentgraph = "false";
    }
    if (getE("saveChatYes").checked == true) {
      saveChat = "true";
    } else {
      saveChat = "false";
    }
  } else if (provider == "together") {
    key = getE("togetherKey").value;
    if (key == "") {
      alert("You must provide an API key.")
      return
    }
    model = getE("togetherModel").value;
    openAIPrompt = getE("togetherAIPrompt").value;
    if (getE("commandYes").checked == true) {
      doCommands = "true";
    } else {
      doCommands = "false";
    }
    if (getE("intentyes").checked == true) {
      intentgraph = "true";
    } else {
      intentgraph = "false";
    }
    if (getE("saveChatYes").checked == true) {
      saveChat = "true";
    } else {
      saveChat = "false";
    }
  } else if (provider == "houndify") {
    key = getE("houndKey").value;
    if (key == "") {
      alert("You must provide a client key.")
      return
    }
    model = "";
    id = getE("houndID").value;
    if (id == "") {
      alert("You must provide a client ID.")
      return
    }
    intentgraph = "false";
  } else {
    key = "";
    id = "";
    model = "";
    intentgraph = "false";
  }

  var data =
    "provider=" +
    provider +
    "&api_key=" +
    key +
    "&model=" +
    model +
    "&api_id=" +
    id +
    "&intent_graph=" +
    intentgraph +
    "&robot_name=" +
    robotName +
    "&openai_prompt=" +
    openAIPrompt +
    "&save_chat=" +
    saveChat +
    "&commands_enable=" +
    doCommands +
    "&endpoint=" +
    endpoint;
  fetch("/api/set_kg_api?" + data)
    .then((response) => response.text())
    .then((response) => {
      updateSetupStatus(response);
      setConn();
    });
}

function checkConn() {
  connValue = document.getElementById("connSelection").value;
  if (connValue == "ip") {
    document.getElementById("portViz").style.display = "block";
  } else {
    document.getElementById("portViz").style.display = "none";
  }
}

function setConn() {
  updateSetupStatus("Setting connection type (ep or ip)...");
  connValue = document.getElementById("connSelection").value;
  port = document.getElementById("portInput").value;
  if (port == "") {
    port = "443";
  }
  url = "";
  if (connValue == "ep") {
    url = "/api-chipper/use_ep";
  } else {
    url = "/api-chipper/use_ip?port=" + port;
  }
  fetch(url)
    .then((response) => response.text())
    .then((response) => {
      if (response == "") {
        updateSetupStatus("error setting up wire-pod, check the logs");
        document.getElementById("config-options").style.display = "block";
        return;
      } else {
        updateSetupStatus(
          "Setup is complete! Wire-pod has started. Redirecting to main page..."
        );
        setTimeout(function () {
          window.location.href = "/";
        }, 3000);
      }
    });
}

function directToIndex() {
  window.location.href = "/index.html";
}
