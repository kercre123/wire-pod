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

function updateBotSelection(element) {
    let xhr = new XMLHttpRequest();
    xhr.open("GET", "/api/get_bot_json");
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
    xhr.responseType = 'json';
    xhr.send();
    xhr.onload = function() {
        document.getElementById("botEditStatus").innerHTML = "";
        document.getElementById("botDeleteStatus").innerHTML = "";
        var listResponse = xhr.response
        var listNum = 0;
        if (listResponse != null) {
            listNum = Object.keys(listResponse).length
        }
        if (listResponse != null && listNum != 0) {
            var select1 = document.createElement("select");
            document.getElementById(element).innerHTML = ""
            select1.name = element + "bots";
            select1.id = element + "bots"
            for (const name in listResponse)
            {
                console.log(listResponse[name]["esn"])
                var option = document.createElement("option");
                option.value = listResponse[name]["esn"];
                option.text = listResponse[name]["esn"]
                select1.appendChild(option);
            }
            var label1 = document.createElement("label");
            label1.innerHTML = "Choose the bot: "
            label1.htmlFor = element + "bots";
            document.getElementById(element).appendChild(label1).appendChild(select1);
        } else {
            console.log("No bots founda")
            var error3 = document.createElement("p");
            var error4 = document.createElement("p");
            error3.innerHTML = "No bots found, you must add one first"
            error4.innerHTML = "No bots found, you must add one first"
            document.getElementById("botEditForm").innerHTML = "";
            document.getElementById("botEditStatus").innerHTML = "";
            document.getElementById("botDeleteStatus").innerHTML = "";
            document.getElementById("botEditSelect").innerHTML = "";
            document.getElementById("botEditSelect").innerHTML = "";
            document.getElementById("botDeleteSelect").innerHTML = "";
            document.getElementById("botDeleteStatus").appendChild(error3);
            document.getElementById("botEditStatus").appendChild(error4);
        }
    };
}

updateBotSelection("botEditSelect")
updateBotSelection("botDeleteSelect")

// create functions from that html to send post to /api/add_bot
function sendBotAdd() {
    const form = document.getElementById('botAddForm');
    const select1 = document.getElementById('unitsAdd');
    const select2 = document.getElementById('firmwarePrefixAdd');
    var data = "esn=" + form.elements['esnAdd'].value + "&location=" + form.elements['locationAdd'].value + "&units=" + select1.value + "&firmwareprefix=" + select2.value;
    var client = new HttpClient();
    var result = document.getElementById('botAddStatus');
    const resultP = document.createElement('p');
    resultP.textContent =  "Adding...";
    result.innerHTML = '';
    result.appendChild(resultP);
    fetch("/api/add_bot?" + data)
    .then(response => response.text())
    .then((response) => {
        resultP.innerHTML = response
        result.innerHTML = '';
        result.appendChild(resultP);
        updateBotSelection("botEditSelect")
        updateBotSelection("botDeleteSelect")
    })
}

function editBotFormCreate() {
    var botNumber = document.getElementById("botEditSelectbots").selectedIndex;
    var xhr = new XMLHttpRequest();
    xhr.open("GET", "/api/get_bot_json");
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
    xhr.responseType = 'json';
    xhr.send();
    xhr.onload = function() {
        var botInfo = xhr.response[botNumber];
        console.log(botInfo)
        // a form doesnt exist yet, create one
            var form = document.createElement("form");
            form.id = "botEditFormNoDiv";
            form.name = "botEditFormNoDiv";
            var label1 = document.createElement("label");
            label1.innerHTML = "ESN: "
            label1.htmlFor = "esnEdit";
            var input1 = document.createElement("input");
            input1.type = "text";
            input1.name = "esnEdit";
            input1.id = "esnEdit";
            input1.value = botInfo["esn"];
            var label2 = document.createElement("label");
            label2.innerHTML = "Location: "
            label2.htmlFor = "locationEdit";
            var input2 = document.createElement("input");
            input2.type = "text";
            input2.name = "locationEdit";
            input2.id = "locationEdit";
            input2.value = botInfo["location"];
            var label3 = document.createElement("label");
            label3.innerHTML = "Units: "
            label3.htmlFor = "unitsEdit";
            var select1 = document.createElement("select");
            select1.name = "unitsEdit";
            select1.id = "unitsEdit";
            var option1 = document.createElement("option");
            option1.value = "F";
            option1.text = "F";
            var option2 = document.createElement("option");
            option2.value = "C";
            option2.text = "C";
            select1.appendChild(option1);
            select1.appendChild(option2);
            if (botInfo["units"] == "F") {
                select1.selectedIndex = 0;
            } else {
                select1.selectedIndex = 1;
            }
            var label4 = document.createElement("label");
            label4.innerHTML = "Firmware Prefix: "
            label4.htmlFor = "firmwarePrefixEdit";
            var select2 = document.createElement("select");
            select2.name = "firmwarePrefixEdit";
            select2.id = "firmwarePrefixEdit";
            var option3 = document.createElement("option");
            option3.value = "1.6";
            option3.text = "1.6 and above";
            var option4 = document.createElement("option");
            option4.value = "1.5";
            option4.text = "1.0-1.5";
            var option5 = document.createElement("option");
            option5.value = "0.11";
            option5.text = "0.14 and below";
            select2.appendChild(option3);
            select2.appendChild(option4);
            select2.appendChild(option5);
            if (botInfo["is_early_opus"] == false && botInfo["use_play_specific"] == false) {
                select2.selectedIndex = 0;
            } else if (botInfo["is_early_opus"] == false && botInfo["use_play_specific"] == true) {
                select2.selectedIndex = 1;
            } else {
                select2.selectedIndex = 2;
            } 
            var button1 = document.createElement("button");
            button1.type = "button";
            button1.id = "botEditSubmit";
            button1.innerHTML = "Submit";
            button1.onclick = function() {
                sendBotEdit();
            }
            form.appendChild(label1);
            form.appendChild(input1);
            form.appendChild(document.createElement("br"));
            form.appendChild(label2);
            form.appendChild(input2);
            form.appendChild(document.createElement("br"));
            form.appendChild(label3);
            form.appendChild(select1);
            form.appendChild(document.createElement("br"));
            form.appendChild(label4);
            form.appendChild(select2);
            form.appendChild(document.createElement("br"));
            form.appendChild(button1);
            document.getElementById("botEditForm").innerHTML = "";
            document.getElementById("botEditForm").appendChild(form);
            // am i missing anything?
        
    }
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

// send post to /api/edit_bot
function sendBotEdit() {
    var form = document.getElementById("botEditFormNoDiv");
    var botNumber = document.getElementById("botEditSelectbots").selectedIndex + 1;
    var data = "number=" + botNumber + "&esn=" + form.elements["esnEdit"].value + "&location=" + form.elements["locationEdit"].value + "&units=" + form.elements["unitsEdit"].value + "&firmwareprefix=" + form.elements["firmwarePrefixEdit"].value;
    var client = new HttpClient();
    var result = document.getElementById('botEditStatus');
    const resultP = document.createElement('p');
    resultP.textContent =  "Editing...";
    result.innerHTML = '';
    result.appendChild(resultP);
    fetch("/api/edit_bot?" + data)
    .then(response => response.text())
    .then((response) => {
        resultP.innerHTML = response
        result.innerHTML = '';
        result.appendChild(resultP);
        updateBotSelection("botEditSelect")
        updateBotSelection("botDeleteSelect")
    })
}

// create a function that deletes a bot
function deleteSelectedBot() {
    var botNumber = document.getElementById("botDeleteSelectbots").selectedIndex + 1;
    var xhr = new XMLHttpRequest();
        var data = "number=" + botNumber;
        var client = new HttpClient();
        var result = document.getElementById('botDeleteStatus');
        const resultP = document.createElement('p');
        resultP.textContent =  "Deleting...";
        result.innerHTML = '';
        result.appendChild(resultP);
        fetch("/api/remove_bot?" + data)
        .then(response => response.text())
        .then((response) => {
            resultP.innerHTML = response
            result.innerHTML = '';
            result.appendChild(resultP);
            updateBotSelection("botEditSelect")
            updateBotSelection("botDeleteSelect")
        })
    
}