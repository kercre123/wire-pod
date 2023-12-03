function checkLanguage() {
    let xhr = new XMLHttpRequest();
    xhr.open("GET", "/api/get_stt_info");
    xhr.send();
    xhr.onload = function() {
        parsed = JSON.parse(xhr.response)
        if (parsed["sttProvider"] != "vosk" && parsed["sttProvider"] != "whisper.cpp") {
            console.log("stt not vosk/whisper")
          document.getElementById("section-language").style.display = "none"
          document.getElementById("languageSelection").value = "en-US"
        } else {
          document.getElementById("section-language").style.display = "block"
          console.log(parsed["sttLanguage"])
          document.getElementById("languageSelection").value = "en-US"
        }
    }
}

function updateSetupStatus(statusString) {
    setupStatus = document.getElementById("setup-status")
    setupStatus.innerHTML = ""
    setupStatusP = document.createElement("p")
    setupStatusP.innerHTML = statusString
    setupStatus.appendChild(setupStatusP)
}

function sendSetupInfo() {
    document.getElementById("config-options").style.display = "none"
    updateSetupStatus("Initiating setup...")
    language = document.getElementById("languageSelection").value

    // set language first
    var langData = "language=" + language;
    document.getElementById("languageSelectionDiv").style.display = "none"
    fetch("/api/set_stt_info?" + langData)
        .then(response => response.text())
        .then((response) => {
            if (response.includes("success")) {
              updateSetupStatus("Language set successfully.")
              initWeatherAPIKey()
            } else if (response.includes("downloading")) {
              updateSetupStatus("Downloading language model...")
              inte = setInterval(function(){
                fetch("/api/get_download_status")
                  .then(response => response.text())
                  .then((response => {
                    statusText = response
                    if (response.includes("success")) {
                      updateSetupStatus("Language set successfully.")
                      initWeatherAPIKey()
                      clearInterval(inte)
                    } else if (response.includes("error")) {
                        document.getElementById("config-options").style.display = "block"
                    } else if (response.includes("not downloading")) {
                      statusText = "Initiating language model download..."
                    }
                    updateSetupStatus(statusText)
                  }))
              }, 500)
            } else if (response.includes("vosk")) {
                initWeatherAPIKey()
            } else if (response.includes("error")) {
              updateSetupStatus(response)
              document.getElementById("config-options").style.display = "block"
              return
            }

        })
}

function initWeatherAPIKey() {
    var provider = document.getElementById("weatherProvider").value;
    if (provider != "") {
    updateSetupStatus("Setting weather API key...")
    var form = document.getElementById("weatherAPIAddForm");

    var data = "provider=" + provider + "&api_key=" + form.elements["apiKey"].value;
    fetch("/api/set_weather_api?" + data)
        .then(response => response.text())
        .then((response) => {
            updateSetupStatus(response)
            initKGAPIKey()
        })
    } else {
        initKGAPIKey()
    }
} 

function initKGAPIKey() {
    updateSetupStatus("Setting knowledge graph settings...")
    var provider = document.getElementById("kgProvider").value
    var key = ""
    var id = ""
    var intentgraph = ""
    var robotName = ""

    if (provider == "openai") {
        key = document.getElementById("openAIKey").value
        if (document.getElementById("intentyes").checked == true) {
            intentgraph = "true"
            robotName = document.getElementById("openAIRobotName").value
        } else {
            intentgraph = "false"
        }
    } else if (provider == "houndify") {
        key = document.getElementById("houndKey").value
        id = document.getElementById("houndID").value
        intentgraph = "false"
    } else {
        key = ""
        id = ""
        intentgraph = "false"
    }

    var data = "provider=" + provider + "&api_key=" + key + "&api_id=" + id + "&intent_graph=" + intentgraph + "&robot_name=" + robotName
    fetch("/api/set_kg_api?" + data)
        .then(response => response.text())
        .then((response) => {
            updateSetupStatus(response)
            setConn()
        })
}

function checkConn() {
    connValue = document.getElementById("connSelection").value
    if (connValue == "ip") {
        document.getElementById("portViz").style.display = "block"
    } else {
        document.getElementById("portViz").style.display = "none"
    }
}

function setConn() {
    updateSetupStatus("Setting connection type (ep or ip)...")
    connValue = document.getElementById("connSelection").value
    port = document.getElementById("portInput").value
    if (port == "") {
        port = "443"
    }
    url = ""
    if (connValue == "ep") {
        url = "/api-chipper/use_ep"
    } else {
        url = "/api-chipper/use_ip?port=" + port
    }
        fetch(url)
        .then(response => response.text())
        .then((response) => {
            if (response == "") {
                updateSetupStatus("error setting up wire-pod, check the logs")
                document.getElementById("config-options").style.display = "block"
                return
            } else {
                updateSetupStatus("Setup is complete! Wire-pod has started. Redirecting to main page...")
                setTimeout(function(){window.location.href = "/";}, 3000)
            }
        })
}

function directToIndex() {
    window.location.href = "/index.html"
}