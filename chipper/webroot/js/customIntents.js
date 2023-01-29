var eventListenerAdded = false

intentsJson = '["intent_greeting_hello", "intent_names_ask", "intent_imperative_eyecolor", "intent_character_age", "intent_explore_start", "intent_system_charger", "intent_system_sleep", "intent_greeting_goodmorning", "intent_greeting_goodnight", "intent_greeting_goodbye", "intent_seasonal_happynewyear", "intent_seasonal_happyholidays", "intent_amazon_signin", "intent_amazon_signin", "intent_imperative_forward", "intent_imperative_turnaround", "intent_imperative_turnleft", "intent_imperative_turnright", "intent_play_rollcube", "intent_play_popawheelie", "intent_play_fistbump", "intent_play_blackjack", "intent_imperative_affirmative", "intent_imperative_negative", "intent_photo_take_extend", "intent_imperative_praise", "intent_imperative_abuse", "intent_weather_extend", "intent_imperative_apologize", "intent_imperative_backup", "intent_imperative_volumedown", "intent_imperative_volumeup", "intent_imperative_lookatme", "intent_imperative_volumelevel_extend", "intent_imperative_shutup", "intent_names_username_extend", "intent_imperative_come", "intent_imperative_love", "intent_knowledge_promptquestion", "intent_clock_checktimer", "intent_global_stop_extend", "intent_clock_settimer_extend", "intent_clock_time", "intent_imperative_quiet", "intent_imperative_dance", "intent_play_pickupcube", "intent_imperative_fetchcube", "intent_imperative_findcube", "intent_play_anytrick", "intent_message_recordmessage_extend", "intent_message_playmessage_extend", "intent_blackjack_hit", "intent_blackjack_stand", "intent_play_keepaway"]'

function updateIntentSelection(element) {
let xhr = new XMLHttpRequest();
xhr.open("GET", "/api/get_custom_intents_json");
xhr.setRequestHeader("Content-Type", "application/json");
xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
xhr.responseType = 'json';
xhr.send();
xhr.onload = function() {
    document.getElementById("editIntentStatus").innerHTML = "";
    document.getElementById("deleteIntentStatus").innerHTML = "";
  var listResponse = xhr.response
  var listNum = 0;
  if (listResponse != null) {
    listNum = Object.keys(listResponse).length
  }
  if (listResponse != null && listNum != 0) {
    
  console.log(listNum)
  var select = document.createElement("select");
  document.getElementById(element).innerHTML = ""
  select.name = element + "intents";
  select.id = element + "intents"
  for (const name in listResponse)
  {
    if (false==listResponse[name]["issystem"]) {
        console.log(listResponse[name]["name"])
        var option = document.createElement("option");
        option.value = listResponse[name]["name"];
        option.text = listResponse[name]["name"]
        select.appendChild(option);
    }
  }
  var label = document.createElement("label");
  label.innerHTML = "Choose the intent: "
  label.htmlFor = element + "intents";
  document.getElementById(element).appendChild(label).appendChild(select);
} else {
    console.log("No intents founda")
    var error1 = document.createElement("p");
    var error2 = document.createElement("p");
    error1.innerHTML = "No intents found, you must add one first"
    error2.innerHTML = "No intents found, you must add one first"
    document.getElementById("editIntentForm").innerHTML = "";
    document.getElementById("editIntentStatus").innerHTML = "";
    document.getElementById("deleteIntentStatus").innerHTML = "";
    document.getElementById("editSelect").innerHTML = "";
    document.getElementById("deleteSelect").innerHTML = "";
    document.getElementById("deleteIntentStatus").appendChild(error1);
    document.getElementById("editIntentStatus").appendChild(error2);
}
};
}

// function that creates select from intentsJson
function createIntentSelect(element) {
    var select = document.createElement("select");
    document.getElementById(element).innerHTML = ""
    select.name = element + "intents";
    select.id = element + "intents"
    var intents = JSON.parse(intentsJson)
    for (const name in intents)
    {
        var option = document.createElement("option");
        option.value = intents[name];
        option.text = intents[name]
        select.appendChild(option);
    }
    var label = document.createElement("label");
    label.innerHTML = "Intent to send to robot after script executed:"
    label.htmlFor = element + "intents";
    document.getElementById(element).appendChild(label).appendChild(select);
}

// get intent from editSelect element and create a form in div#editIntentForm to edit it
// options are: name, description, utterances, intent, paramname, paramvalue, exec
function editFormCreate() {
    var intentNumber = document.getElementById("editSelectintents").selectedIndex;
    var xhr = new XMLHttpRequest();
    xhr.open("GET", "/api/get_custom_intents_json");
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
    xhr.responseType = 'json';
    xhr.send();
    xhr.onload = function() {
        var intentResponse = xhr.response[intentNumber];
        if (intentResponse != null) {
            console.log(intentResponse)
            var form = document.createElement("form");
            form.id = "editIntentForm";
            form.name = "editIntentForm";
            var name = document.createElement("input");
            name.type = "text";
            name.name = "name";
            name.id = "name";
            // create label for name
            var nameLabel = document.createElement("label");
            nameLabel.innerHTML = "Name: "
            nameLabel.htmlFor = "name";
            name.value = intentResponse["name"];
            var description = document.createElement("input");
            description.type = "text";
            description.name = "description";
            description.id = "description";
            // create label for description
            var descriptionLabel = document.createElement("label");
            descriptionLabel.innerHTML = "Description: "
            descriptionLabel.htmlFor = "description";
            description.value = intentResponse["description"];
            var utterances = document.createElement("input");
            utterances.type = "text";
            utterances.name = "utterances";
            utterances.id = "utterances";
            // create label for utterances
            var utterancesLabel = document.createElement("label");
            utterancesLabel.innerHTML = "Utterances: "
            utterancesLabel.htmlFor = "utterances";
            utterances.value = intentResponse["utterances"];
            var intent = document.createElement("select");
            var intents = JSON.parse(intentsJson)
            for (const name in intents)
            {
                var option = document.createElement("option");
                option.value = intents[name];
                option.text = intents[name]
                intent.appendChild(option);
            }
            intent.type = "text";
            intent.name = "intent";
            intent.id = "intent";
            // create label for intent
            var intentLabel = document.createElement("label");
            intentLabel.innerHTML = "Intent: "
            intentLabel.htmlFor = "intent";
            intent.value = intentResponse["intent"];
            var paramname = document.createElement("input");
            paramname.type = "text";
            paramname.name = "paramname";
            paramname.id = "paramname";
            // create label for paramname
            var paramnameLabel = document.createElement("label");
            paramnameLabel.innerHTML = "Param Name: "
            paramnameLabel.htmlFor = "paramname";
            paramname.value = intentResponse["params"]["paramname"];
            var paramvalue = document.createElement("input");
            paramvalue.type = "text";
            paramvalue.name = "paramvalue";
            paramvalue.id = "paramvalue";
            // create label for paramvalue
            var paramvalueLabel = document.createElement("label");
            paramvalueLabel.innerHTML = "Param Value: "
            paramvalueLabel.htmlFor = "paramvalue";
            paramvalue.value = intentResponse["params"]["paramvalue"];
            var exec = document.createElement("input");
            exec.type = "text";
            exec.name = "exec";
            exec.id = "exec";
            // create label for exec
            var execLabel = document.createElement("label");
            execLabel.innerHTML = "Exec: "
            execLabel.htmlFor = "exec";
            exec.value = intentResponse["exec"];
            // execargs
            var execargs = document.createElement("input");
            execargs.type = "text";
            execargs.name = "execargs";
            execargs.id = "execargs";
            // create label for execargs
            var execargsLabel = document.createElement("label");
            execargsLabel.innerHTML = "Exec Args: "
            execargsLabel.htmlFor = "execargs";
            execargs.value = intentResponse["execargs"];
            // create button that launches function
            var submit = document.createElement("button");
            submit.type = "button";
            submit.id = "submit";
            submit.innerHTML = "Submit";
            submit.onclick = function() {
                editIntent(intentNumber);
            }
            form.appendChild(nameLabel).appendChild(name);
            form.appendChild(document.createElement("br"));
            form.appendChild(descriptionLabel).appendChild(description);
            form.appendChild(document.createElement("br"));
            form.appendChild(utterancesLabel).appendChild(utterances);
            form.appendChild(document.createElement("br"));
            form.appendChild(intentLabel).appendChild(intent);
            form.appendChild(document.createElement("br"));
            form.appendChild(paramnameLabel).appendChild(paramname);
            form.appendChild(document.createElement("br"));
            form.appendChild(paramvalueLabel).appendChild(paramvalue);
            form.appendChild(document.createElement("br"));
            form.appendChild(execLabel).appendChild(exec);
            form.appendChild(document.createElement("br"));
            form.appendChild(execargsLabel).appendChild(execargs);
            form.appendChild(document.createElement("br"));
            form.appendChild(submit);
            document.getElementById("editIntentForm").innerHTML = "";
            document.getElementById("editIntentForm").appendChild(form);
        } else {
            console.log("No intents founda")
            var error1 = document.createElement("p");
            var error2 = document.createElement("p");
            error1.innerHTML = "No intents found, you must add one first"
            error2.innerHTML = "No intents found, you must add one first"
            document.getElementById("editIntentForm").innerHTML = "";
            document.getElementById("editIntentStatus").innerHTML = "";
            document.getElementById("deleteIntentStatus").innerHTML = "";
            document.getElementById("editSelect").innerHTML = "";
            document.getElementById("deleteSelect").innerHTML = "";
            document.getElementById("deleteIntentStatus").appendChild(error1);
            document.getElementById("editIntentStatus").appendChild(error2);
        }
    };
}

// create editIntent function that sends post to /api/edit_custom_intent, get index of intent to edit
// form data should include the intent number, name, description, utterances, intent, paramname, paramvalue, exec
function editIntent(intentNumber) {
    console.log(intentNumber)
        var formData = new FormData();
        formData.append("number", intentNumber+1);
        formData.append("name", document.getElementById("name").value);
        formData.append("description", document.getElementById("description").value);
        formData.append("utterances", document.getElementById("utterances").value);
        formData.append("intent", document.getElementById("intent").value);
        formData.append("paramname", document.getElementById("paramname").value);
        formData.append("paramvalue", document.getElementById("paramvalue").value);
        formData.append("exec", document.getElementById("exec").value);
        formData.append("execargs", document.getElementById("execargs").value);
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "/api/edit_custom_intent");
        xhr.send(formData);
        xhr.onload = function() {
            var response = xhr.response;
            console.log(response);
                console.log("Intent edited")
                var success = document.createElement("p");
                success.innerHTML = "Intent edited"
                document.getElementById("editIntentStatus").innerHTML = "";
                document.getElementById("editIntentStatus").appendChild(success);
        }
    
}

updateIntentSelection("editSelect")
updateIntentSelection("deleteSelect")
createIntentSelect("intentAddSelect")

var HttpClient = function() {
    this.get = function(aUrl, aCallback) {
        var anHttpRequest = new XMLHttpRequest();
        anHttpRequest.onreadystatechange = function() { 
            if (anHttpRequest.readyState == 4 && anHttpRequest.status == 200)
                aCallback(anHttpRequest.responseText);
        }

        anHttpRequest.open( "GET", aUrl, true );            
        anHttpRequest.send( null );
    }
}

function deleteSelectedIntent() {
    var intentNumber = document.getElementById("deleteSelectintents").selectedIndex;
    var formData = new FormData();
    formData.append("number", intentNumber+1);
    console.log(intentNumber+1)
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/api/remove_custom_intent");
    xhr.send(formData);
    xhr.onload = function() {
        var response = xhr.response;
        console.log(response);
        console.log("Intent deleted")
        var success = document.createElement("p");
        success.innerHTML = "Intent deleted"
        document.getElementById("deleteIntentStatus").innerHTML = "";
        document.getElementById("deleteIntentStatus").appendChild(success);
        updateIntentSelection("editSelect")
        updateIntentSelection("deleteSelect")
    }
}

function sendIntentAdd() {
    const form = document.getElementById('intentAddForm');
    var data = "name=" + form.elements['nameAdd'].value + "&description=" + form.elements['descriptionAdd'].value + "&utterances=" + form.elements['utterancesAdd'].value + "&intent=" + form.elements['intentAddSelectintents'].value + "&paramname=" + form.elements['paramnameAdd'].value + "&paramvalue=" + form.elements['paramvalueAdd'].value + "&exec=" + form.elements['execAdd'].value + "&execargs=" + form.elements['execAddArgs'].value;
    var client = new HttpClient();
    var result = document.getElementById('addIntentStatus');
    const resultP = document.createElement('p');
    resultP.textContent =  "Adding...";
    result.innerHTML = '';
    result.appendChild(resultP);
    fetch("/api/add_custom_intent?" + data)
    .then(response => response.text())
    .then((response) => {
        resultP.innerHTML = response
        result.innerHTML = '';
        result.appendChild(resultP);
        updateIntentSelection("editSelect")
        updateIntentSelection("deleteSelect")
    })
}

function sendLinkForm() {
    var data = "esn=" + document.getElementById("link_esn").value + "&target=" + document.getElementById("link_ip").value
    console.log(data)
    fetch("/link-esn-and-target?" + data)
    .then(response => response.text())
    .then((response) => {
        if (response.includes("success")) {
            alert("Bot successfully linked! You may now set your bot up via https://keriganc.com/vector-epod-setup")
        } else {
            alert(response)
        }
    })
}

// ##############################################
function toggleContent(element) {
  if (element.style.display === "block") {
    element.style.display = "none";
  } else {
    element.style.display = "block";
  }
}

var headings = document.getElementsByTagName("h2");
for (var i = 0; i < headings.length; i++) {
  headings[i].addEventListener("click", function() {
    toggleContent(this.nextElementSibling);
  });
}

function togglePlusMinusSymbols() {
  var h2Elements = document.getElementsByTagName("h2");
  for (var i = 0; i < h2Elements.length; i++) {
    h2Elements[i].addEventListener("click", function() {
      var plusMinusElement = this.firstElementChild;
      if (plusMinusElement.innerHTML === "-") {
        plusMinusElement.innerHTML = "+";
      } else {
        plusMinusElement.innerHTML = "-";
      }
    });
  }
}
togglePlusMinusSymbols();

GetLog = false

function showLog() {
    document.getElementById("section-intents").style.display = "none";
    document.getElementById("section-log").style.display = "block";
    document.getElementById("botTranscriptedText").style.display = "block";
    GetLog = true
    logDiv = document.getElementById("botTranscriptedText")
    logP = document.createElement("p")
    setInterval(function() {
        if (GetLog == false) {
            return
        }
        let xhr = new XMLHttpRequest();
        xhr.open("GET", "/api/get_logs");
        xhr.send();
        xhr.onload = function() {
            logDiv.innerHTML = ""
            if (xhr.response == "") {
                logP.innerHTML = "No logs yet, you must say a command to Vector."
            } else {
                logP.innerHTML = xhr.response
            }
            logDiv.appendChild(logP)
        }
    }, 1000)
}

function showIntents() {
    GetLog = false
    document.getElementById("section-log").style.display = "none";
    document.getElementById("section-intents").style.display = "block";
}

function showBotConfig() {
    GetLog = false
    document.getElementById("section-log").style.display = "none";
    document.getElementById("section-intents").style.display = "none";
}

function showWeather() {
    document.getElementById("section-weather").style.display = "block";
    document.getElementById("section-stt").style.display = "none";
    document.getElementById("section-restart").style.display = "none";
    document.getElementById("section-kg").style.display = "none";
}

function showKG() {
    document.getElementById("section-weather").style.display = "none";
    document.getElementById("section-stt").style.display = "none";
    document.getElementById("section-restart").style.display = "none";
    document.getElementById("section-kg").style.display = "block";
}

function showSTT() {
    document.getElementById("section-weather").style.display = "none";
    document.getElementById("section-kg").style.display = "none";
    document.getElementById("section-stt").style.display = "block";
    document.getElementById("section-restart").style.display = "none";
}

function showRestart() {
    document.getElementById("section-weather").style.display = "none";
    document.getElementById("section-kg").style.display = "none";
    document.getElementById("section-stt").style.display = "none";
    document.getElementById("section-restart").style.display = "block";
}

function checkWeather() {
    if (document.getElementById("weatherProvider").value=="") {
        document.getElementById("apiKey").value = "";
        document.getElementById("apiKeySpan").style.display = "none";
    }
    else {
        document.getElementById("apiKeySpan").style.display = "block";
    }
}

function sendWeatherAPIKey(element) {
    var form = document.getElementById("weatherAPIAddForm");
    var provider = document.getElementById("weatherProvider").value;

    var data = "provider=" + provider + "&api_key=" + form.elements["apiKey"].value;
    var result = document.getElementById('addWeatherProviderAPIStatus');
    const resultP = document.createElement('p');
    resultP.textContent =  "Saving...";
    result.innerHTML = '';
    result.appendChild(resultP);
    fetch("/api/set_weather_api?" + data)
        .then(response => response.text())
        .then((response) => {
            resultP.innerHTML = response
            result.innerHTML = '';
            result.appendChild(resultP);
        })
}

function updateWeatherAPI() {
    fetch("/api/get_weather_api")
        .then(response => response.text())
        .then((response) => {
            obj = JSON.parse(response);
            document.getElementById("weatherProvider").value = obj.weatherProvider;
            document.getElementById("apiKey").value = obj.weatherApiKey;
            checkWeather();
        })
}

function checkKG() {
    if (document.getElementById("kgProvider").value=="") {
        document.getElementById("kgKey").value = "";
        document.getElementById("kgKeySpan").style.display = "none";
    }
    else {
        document.getElementById("kgKeySpan").style.display = "block";
    }
}

function sendKGAPIKey(element) {
    var form = document.getElementById("kgAPIAddForm");
    var provider = document.getElementById("kgProvider").value;

    var data = "provider=" + provider + "&api_key=" + form.elements["kgKey"].value;
    var result = document.getElementById('addKGProviderAPIStatus');
    const resultP = document.createElement('p');
    resultP.textContent =  "Saving...";
    result.innerHTML = '';
    result.appendChild(resultP);
    fetch("/api/set_kg_api?" + data)
        .then(response => response.text())
        .then((response) => {
            resultP.innerHTML = response
            result.innerHTML = '';
            result.appendChild(resultP);
        })
}

function updateKGAPI() {
    fetch("/api/get_kg_api")
        .then(response => response.text())
        .then((response) => {
            obj = JSON.parse(response);
            document.getElementById("kgProvider").value = obj.kgProvider;
            document.getElementById("kgKey").value = obj.kgApiKey;
            checkKG();
        })
}

function sendSTTLanguage() {
    var language = document.getElementById("sttLanguage").value;
    var data = "language=" + language;
    var result = document.getElementById('addSTTStatus');
    const resultP = document.createElement('p');
    resultP.textContent =  "Saving...";
    result.innerHTML = '';
    result.appendChild(resultP);
    fetch("/api/set_stt_info?" + data)
        .then(response => response.text())
        .then((response) => {
            resultP.innerHTML = response
            result.innerHTML = '';
            result.appendChild(resultP);
        })
}

function updateSTTLanguage() {
    fetch("/api/get_stt_info")
        .then(response => response.text())
        .then((response) => {
            obj = JSON.parse(response);
            //document.getElementById("sttProvider").value = obj.sttProvider;
            document.getElementById("sttLanguage").value = obj.sttLanguage;
        })
}

function sendRestart() {
    fetch("/api/reset")
        .then(response => response.text())
        .then((response) => {
            resultP.innerHTML = response
            result.innerHTML = '';
            result.appendChild(resultP);
        })
}

