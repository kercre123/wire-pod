const intentsJson = JSON.parse(
  '["intent_greeting_hello", "intent_names_ask", "intent_imperative_eyecolor", "intent_character_age", "intent_explore_start", "intent_system_charger", "intent_system_sleep", "intent_greeting_goodmorning", "intent_greeting_goodnight", "intent_greeting_goodbye", "intent_seasonal_happynewyear", "intent_seasonal_happyholidays", "intent_amazon_signin", "intent_imperative_forward", "intent_imperative_turnaround", "intent_imperative_turnleft", "intent_imperative_turnright", "intent_play_rollcube", "intent_play_popawheelie", "intent_play_fistbump", "intent_play_blackjack", "intent_imperative_affirmative", "intent_imperative_negative", "intent_photo_take_extend", "intent_imperative_praise", "intent_imperative_abuse", "intent_weather_extend", "intent_imperative_apologize", "intent_imperative_backup", "intent_imperative_volumedown", "intent_imperative_volumeup", "intent_imperative_lookatme", "intent_imperative_volumelevel_extend", "intent_imperative_shutup", "intent_names_username_extend", "intent_imperative_come", "intent_imperative_love", "intent_knowledge_promptquestion", "intent_clock_checktimer", "intent_global_stop_extend", "intent_clock_settimer_extend", "intent_clock_time", "intent_imperative_quiet", "intent_imperative_dance", "intent_play_pickupcube", "intent_imperative_fetchcube", "intent_imperative_findcube", "intent_play_anytrick", "intent_message_recordmessage_extend", "intent_message_playmessage_extend", "intent_blackjack_hit", "intent_blackjack_stand", "intent_play_keepaway"]'
);

var GetLog = false;
let reminderCounter = 0; 

const getE = (element) => document.getElementById(element);

function updateIntentSelection(element) {
  fetch("/api/get_custom_intents_json")
    .then((response) => response.json())
    .then((listResponse) => {
      const container = getE(element);
      container.innerHTML = "";
      if (listResponse && listResponse.length > 0) {
        const select = document.createElement("select");
        select.name = `${element}intents`;
        select.id = `${element}intents`;
        listResponse.forEach((intent) => {
          if (!intent.issystem) {
            const option = document.createElement("option");
            option.value = intent.name;
            option.text = intent.name;
            select.appendChild(option);
          }
        });
        const label = document.createElement("label");
        label.innerHTML = "Choose the intent: ";
        label.htmlFor = `${element}intents`;
        container.appendChild(label).appendChild(select);

        select.addEventListener("change", hideEditIntents);
      } else {
        const error = document.createElement("p");
        error.innerHTML = "No intents found, you must add one first";
        container.appendChild(error);
      }
    }).catch(() => {
      // Do nothing
    });
}

function checkInited() {
  fetch("/api/is_api_v3").then((response) => {
    if (!response.ok) {
      alert(
        "This webroot does not match with the wire-pod binary. Some functionality will be broken. There was either an error during the last update, or you did not precisely follow the update guide. https://github.com/kercre123/wire-pod/wiki/Things-to-Know#updating-wire-pod"
      );
    }
  });

  fetch("/api/get_config")
    .then((response) => response.json())
    .then((config) => {
      if (!config.pastinitialsetup) {
        window.location.href = "/initial.html";
      }
    });
}

function createIntentSelect(element) {
  const select = document.createElement("select");
  select.name = `${element}intents`;
  select.id = `${element}intents`;
  intentsJson.forEach((intent) => {
    const option = document.createElement("option");
    option.value = intent;
    option.text = intent;
    select.appendChild(option);
  });
  const label = document.createElement("label");
  label.innerHTML = "Intent to send to robot after script executed:";
  label.htmlFor = `${element}intents`;
  getE(element).innerHTML = "";
  getE(element).appendChild(label).appendChild(select);
}

function editFormCreate() {
  const intentNumber = getE("editSelectintents").selectedIndex;

  fetch("/api/get_custom_intents_json")
    .then((response) => response.json())
    .then((intents) => {
      const intent = intents[intentNumber];
      if (intent) {
        const form = document.createElement("form");
        form.id = "editIntentForm";
        form.name = "editIntentForm";
        form.innerHTML = `
          <label for="name">Name:<br><input type="text" id="name" value="${intent.name}"></label><br>
          <label for="description">Description:<br><input type="text" id="description" value="${intent.description}"></label><br>
          <label for="utterances">Utterances:<br><input type="text" id="utterances" value="${intent.utterances.join(",")}"></label><br>
          <label for="intent">Intent:<br><select id="intent">${intentsJson
            .map(
              (name) =>
                `<option value="${name}" ${name === intent.intent ? "selected" : ""
                }>${name}</option>`
            )
            .join("")}</select></label><br>
          <label for="paramname">Param Name:<br><input type="text" id="paramname" value="${intent.params.paramname}"></label><br>
          <label for="paramvalue">Param Value:<br><input type="text" id="paramvalue" value="${intent.params.paramvalue}"></label><br>
          <label for="exec">Exec:<br><input type="text" id="exec" value="${intent.exec}"></label><br>
          <label for="execargs">Exec Args:<br><input type="text" id="execargs" value="${intent.execargs.join(",")}"></label><br>
          <label for="luascript">Lua code to run:</label><br><textarea id="luascript">${intent.luascript}</textarea>
          <button onclick="editIntent(${intentNumber})">Submit</button>
        `;
        //form.querySelector("#submit").onclick = () => editIntent(intentNumber);
        getE("editIntentForm").innerHTML = "";
        getE("editIntentForm").appendChild(form);
        showEditIntents();
      } else {
        displayError("editIntentForm", "No intents found, you must add one first");
      }
    }).catch((error) => {
      console.error(error);
      displayError("editIntentForm", "Error fetching intents");
    })
}

function editIntent(intentNumber) {
  const data = {
    number: intentNumber + 1,
    name: getE("name").value,
    description: getE("description").value,
    utterances: getE("utterances").value.split(","),
    intent: getE("intent").value,
    params: {
      paramname: getE("paramname").value,
      paramvalue: getE("paramvalue").value,
    },
    exec: getE("exec").value,
    execargs: getE("execargs").value.split(","),
    luascript: getE("luascript").value,
  };

  fetch("/api/edit_custom_intent", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("editIntentStatus", response);
      alert(response)
      updateIntentSelection("editSelect");
      updateIntentSelection("deleteSelect");
    });
}

function deleteSelectedIntent() {
  const intentNumber = getE("editSelectintents").selectedIndex + 1;

  fetch("/api/remove_custom_intent", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ number: intentNumber }),
  })
    .then((response) => response.text())
    .then((response) => {
      hideEditIntents();
      alert(response)
      updateIntentSelection("editSelect");
      updateIntentSelection("deleteSelect");
    });
}

function sendIntentAdd() {
  const form = getE("intentAddForm");
  const data = {
    name: form.elements["nameAdd"].value,
    description: form.elements["descriptionAdd"].value,
    utterances: form.elements["utterancesAdd"].value.split(","),
    intent: form.elements["intentAddSelectintents"].value,
    params: {
      paramname: form.elements["paramnameAdd"].value,
      paramvalue: form.elements["paramvalueAdd"].value,
    },
    exec: form.elements["execAdd"].value,
    execargs: form.elements["execAddArgs"].value.split(","),
    luascript: form.elements["luaAdd"].value,
  };
  if (!data.name || !data.description || !data.utterances) {
    displayMessage("addIntentStatus", "A required input is missing. You need a name, description, and utterances.");
    alert("A required input is missing. You need a name, description, and utterances.")
    return
  }

  displayMessage("addIntentStatus", "Adding...");

  fetch("/api/add_custom_intent", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addIntentStatus", response);
      alert(response)
      updateIntentSelection("editSelect");
      updateIntentSelection("deleteSelect");
    });
}

function checkWeather() {
  getE("apiKeySpan").style.display = getE("weatherProvider").value ? "block" : "none";
}

function sendWeatherAPIKey() {
  const data = {
    provider: getE("weatherProvider").value,
    key: getE("apiKey").value,
  };

  displayMessage("addWeatherProviderAPIStatus", "Saving...");

  fetch("/api/set_weather_api", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addWeatherProviderAPIStatus", response);
    });
}

function updateWeatherAPI() {
  fetch("/api/get_weather_api")
    .then((response) => response.json())
    .then((data) => {
      getE("weatherProvider").value = data.provider;
      getE("apiKey").value = data.key;
      checkWeather();
    });
}

function populateRobotList() {
  return fetch("/api-sdk/get_sdk_info")
    .then((response) => {
        if (!response.ok) return Promise.resolve(); 
        return response.json();
    })
    .then((jsonResp) => {
        const botList = getE("targetBot");
        
        if (jsonResp && jsonResp["robots"]) {
            for (var i = 0; i < jsonResp["robots"].length; i++) {
                let exists = false;
                for (let j = 0; j < botList.options.length; j++) {
                    if (botList.options[j].value === jsonResp["robots"][i]["esn"]) {
                        exists = true;
                        break;
                    }
                }
                if (!exists) {
                    var option = document.createElement("option");
                    option.text = jsonResp["robots"][i]["esn"];
                    option.value = jsonResp["robots"][i]["esn"];
                    botList.add(option);
                }
            }
        }
    })
    .catch((error) => {
        console.error('Unable to get SDK info:', error);
    });
}

function checkProductivity() {
  const isTodoist = getE("todoistEnable").checked;
  getE("productivityKeySpan").style.display = isTodoist ? "block" : "none";
}

function toggleManualReminders() {
   const enabled = getE("enableManualReminders").checked;
   getE("manualRemindersWrapper").style.display = enabled ? "block" : "none";
   getE("manualAddBtn").style.display = enabled ? "block" : "none";
}

function toggleAccordion(id) {
    const content = document.getElementById(id + "_content");
    const header = document.getElementById(id + "_header");
    
    if (content.classList.contains("show")) {
        content.classList.remove("show");
        header.classList.remove("active");
    } else {
        content.classList.add("show");
        header.classList.add("active");
    }
}

function updateReminderTitle(id, value) {
    document.getElementById(id + "_title").innerText = value || "New Reminder";
}

function addReminderBlock(data = null) {
  reminderCounter++;
  const id = `rem_${reminderCounter}`;
  const container = getE("manualRemindersContainer");

  const block = document.createElement("div");
  block.className = "reminder-block";
  block.id = id;

  const reminderName = data ? data.id : "reminder" + Math.random().toString(36).substring(2, 10);
  const displayTitle = data ? data.id : "New Reminder";
  const reminderImage = data ? data.image : "";
  const scheduleType = data && data.schedule ? data.schedule.type : "daily";
  const requireConfirm = data && data.require_confirmation === true ? "checked" : "";
  const isEnabled = data && data.enabled === false ? "" : "checked";
  const snoozeVal = data && data.snooze_minutes ? data.snooze_minutes : 10;

  block.innerHTML = `
    <div class="accordion-header" id="${id}_header" onclick="toggleAccordion('${id}')">
      <span id="${id}_title">${displayTitle}</span>
      <i class="fa-solid fa-chevron-down" style="float:right;"></i>
    </div>
    
    <div class="accordion-content" id="${id}_content">
        <div class="reminder-grid">
            <div class="reminder-section-divider">Identification & State</div>
            
            <label for="${id}_enabled">Status</label>
            <div style="display:flex; align-items:center;">
                <input type="checkbox" class="reminder-enabled" id="${id}_enabled" ${isEnabled}>
                <label for="${id}_enabled" class="checkbox-label" style="font-weight:bold; color:var(--fg-color); margin-left:10px;">Enabled</label>
            </div>

            <label>ID / Name</label>
            <input type="text" class="tinput reminder-id-val" value="${reminderName}" oninput="updateReminderTitle('${id}', this.value)" placeholder="e.g. meds_morning">

            <div class="reminder-section-divider">Behaviour</div>

            <label>Confirmation</label>
            <div style="display:flex; align-items:center;">
                <input type="checkbox" class="reminder-req-confirm" id="${id}_confirm" ${requireConfirm}>
                <label for="${id}_confirm" class="checkbox-label" style="margin-left:10px;">Requires verbal "Yes" (otherwise snooze)</label>
            </div>

            <label>Snooze Time (Minutes)</label>
            <input type="number" class="tinput reminder-snooze-val" value="${snoozeVal}" min="1" placeholder="10">

            <label>Image</label>
            <div>
                <input type="hidden" class="reminder-img-existing" value="${reminderImage}">
                ${reminderImage ? `
                    <div style="margin-bottom: 8px; display: flex; align-items: center; gap: 15px;">
                        <img src="/api/productivity-images/${reminderImage}" style="max-width: 80px; border: 1px solid #555; border-radius: 4px;">
                        <small style="color:gray;">Current: ${reminderImage}</small>
                    </div>
                ` : ''}
                <input type="file" class="reminder-file-input" accept="image/png">
            </div>

            <label>Phrases</label>
            <div>
                <div class="phrases-container" id="${id}_phrases"></div>
                <button type="button" class="btn-generic" style="margin-top:10px;" onclick="addPhraseInput('${id}_phrases')">+ Add Phrase</button>
            </div>

            <div class="reminder-section-divider">Scheduling</div>

            <label>Repeat Type</label>
            <select class="reminder-schedule-type" onchange="toggleScheduleType('${id}', this.value)">
              <option value="daily" ${scheduleType === 'daily' ? 'selected' : ''}>Daily (Specific Time)</option>
              <option value="hourly" ${scheduleType === 'hourly' ? 'selected' : ''}>Hourly</option>
              <option value="random_interval" ${scheduleType === 'random_interval' ? 'selected' : ''}>Random Interval</option>
            </select>

            <div class="reminder-full-width" id="${id}_schedule_options">
            </div>
        </div>

        <div class="reminder-actions" style="margin-top: 25px; border-top: 1px solid #444; padding-top: 15px; display: flex; justify-content: flex-end; gap: 10px;">
            <button type="button" class="btn-generic btn-test" onclick="testReminder('${id}')">Test on Vector</button>
            <button type="button" class="btn-generic btn-remove" onclick="document.getElementById('${id}').remove()">Remove Reminder</button>
        </div>
    </div>
  `;

  container.appendChild(block);

  if (data && data.phrases) {
      data.phrases.forEach(phrase => addPhraseInput(`${id}_phrases`, phrase));
  } else {
      addPhraseInput(`${id}_phrases`);
  }

  toggleScheduleType(id, scheduleType, data ? data.schedule : null);
}

function testReminder(id) {
  const block = document.getElementById(id);
  const targetBot = getE("targetBot").value;
  
  if (!targetBot) {
      alert("Please select a target robot first.");
      return;
  }

  const reminderName = block.querySelector(".reminder-id-val").value;
  const existingImage = block.querySelector(".reminder-img-existing").value;
  const fileInput = block.querySelector(".reminder-file-input");
  const requireConfirm = block.querySelector(".reminder-req-confirm").checked;
  const snoozeMinutes = parseInt(block.querySelector(".reminder-snooze-val").value) || 10;

  const formData = new FormData();
  formData.append("target_robot", targetBot);

  let imageName = existingImage;
  if (fileInput.files.length > 0) {
      imageName = fileInput.files[0].name;
      formData.append("files", fileInput.files[0]);
  }

  const phrases = [];
  block.querySelectorAll(".phrase-val").forEach(input => {
      if(input.value.trim() !== "") phrases.push(input.value.trim());
  });

  const config = {
      id: reminderName,
      image: imageName,
      phrases: phrases,
      require_confirmation: requireConfirm,
      snooze_minutes: snoozeMinutes,
      schedule: { type: "test" } 
  };

  formData.append("reminder_config", JSON.stringify(config));

  displayMessage("addProductivityProviderAPIStatus", "Sending test...");

  fetch("/api/test_productivity_reminder", {
      method: "POST",
      body: formData
  })
  .then(response => response.text())
  .then(text => {
      displayMessage("addProductivityProviderAPIStatus", text);
  })
  .catch(err => {
      displayMessage("addProductivityProviderAPIStatus", "Error: " + err);
  });
}

function addPhraseInput(containerId, value = "") {
  const container = getE(containerId);
  const div = document.createElement("div");
  div.className = "phrase-row";
  div.innerHTML = `
    <input type="text" class="tinput phrase-val" value="${value}" style="flex: 1;" placeholder="Spoken phrase...">
    <button type="button" class="btn-generic btn-remove" style="min-width: 40px;" onclick="this.parentElement.remove()">X</button>
  `;
  container.appendChild(div);
}

function toggleScheduleType(reminderId, type, existingData = null) {
  const container = getE(`${reminderId}_schedule_options`);
  container.innerHTML = "";
  
  const existingDays = existingData && existingData.days ? existingData.days : [];
  const existingHours = existingData && existingData.hours ? existingData.hours : [];

  if (type === "daily") {
    const timeVal = existingData ? existingData.time : "08:00";
    container.innerHTML = `
      <div class="reminder-grid">
        ${getDaysCheckboxHTML(existingDays)}
        <label>Time (HH:MM)</label>
        <input type="time" class="tinput sched-daily-time" value="${timeVal}">
      </div>
    `;
  } else if (type === "hourly") {
     const minVal = existingData ? existingData.minute : "0";
     container.innerHTML = `
       <div class="reminder-grid">
         ${getDaysCheckboxHTML(existingDays)}
         ${getHoursCheckboxHTML(existingHours)}
         <label>Minute past hour</label>
         <input type="number" min="0" max="59" class="tinput sched-hourly-minute" value="${minVal}">
       </div>
     `;
  } else if (type === "random_interval") {
    const minVal = existingData ? existingData.min_minutes : "60";
    const maxVal = existingData ? existingData.max_minutes : "120";
    container.innerHTML = `
      <div class="reminder-grid">
        ${getDaysCheckboxHTML(existingDays)}
        ${getHoursCheckboxHTML(existingHours)}
        <label>Min Minutes</label>
        <input type="number" class="tinput sched-rnd-min" value="${minVal}">
        <label>Max Minutes</label>
        <input type="number" class="tinput sched-rnd-max" value="${maxVal}">
      </div>
    `;
  }
}

function getDaysCheckboxHTML(existingDays = []) {
    const isChecked = (day) => existingDays.includes(day) ? "checked" : "";
    return `
      <label>Active Days</label>
      <div style="display: flex; flex-wrap: wrap; gap: 8px;">
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Mon" ${isChecked('Mon')}> Mon</label>
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Tue" ${isChecked('Tue')}> Tue</label>
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Wed" ${isChecked('Wed')}> Wed</label>
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Thu" ${isChecked('Thu')}> Thu</label>
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Fri" ${isChecked('Fri')}> Fri</label>
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Sat" ${isChecked('Sat')}> Sat</label>
         <label style="color:white; font-size:0.8em;"><input type="checkbox" class="sched-day-check" value="Sun" ${isChecked('Sun')}> Sun</label>
      </div>
    `;
}

function getHoursCheckboxHTML(existingHours = []) {
    let html = `
      <label>Active Hours<br><small>(Optional)</small></label>
      <div class="hour-grid">
    `;
    for (let i = 0; i < 24; i++) {
        const isChecked = existingHours.includes(i) ? "checked" : "";
        html += `
          <div class="hour-item">
            <span class="hour-label">${i}h</span>
            <input type="checkbox" class="sched-hour-check" value="${i}" ${isChecked} style="width:18px; height:18px;">
          </div>
        `;
    }
    html += `</div>`;
    return html;
}

function collectManualConfigData(formDataObj) {
  const enabled = getE("enableManualReminders").checked;
  if (!enabled) return [];

  const blocks = document.querySelectorAll("#manualRemindersContainer .reminder-block");
  const config = [];

  blocks.forEach((block, index) => {
    const id = block.querySelector(".reminder-id-val").value;
    const isEnabled = block.querySelector(".reminder-enabled").checked;
    const existingImage = block.querySelector(".reminder-img-existing").value;
    const fileInput = block.querySelector(".reminder-file-input");
    const requireConfirm = block.querySelector(".reminder-req-confirm").checked;
    const snoozeMinutes = parseInt(block.querySelector(".reminder-snooze-val").value) || 10;

    let imageName = existingImage;
    if (fileInput.files.length > 0) {
        imageName = fileInput.files[0].name;
        formDataObj.append("files", fileInput.files[0]);
    }

    const phrases = [];
    block.querySelectorAll(".phrase-val").forEach(input => {
        if(input.value.trim() !== "") phrases.push(input.value.trim());
    });

    const schedType = block.querySelector(".reminder-schedule-type").value;
    let schedule = { type: schedType };

    const days = [];
    block.querySelectorAll(".sched-day-check:checked").forEach(chk => days.push(chk.value));
    if (days.length > 0) {
        schedule.days = days;
    }

    const hours = [];
    block.querySelectorAll(".sched-hour-check:checked").forEach(chk => hours.push(parseInt(chk.value)));
    if (hours.length > 0) {
        schedule.hours = hours;
    }

    if (schedType === "daily") {
        schedule.time = block.querySelector(".sched-daily-time").value;
    } else if (schedType === "hourly") {
        schedule.minute = parseInt(block.querySelector(".sched-hourly-minute").value) || 0;
    } else {
        schedule.min_minutes = parseInt(block.querySelector(".sched-rnd-min").value) || 60;
        schedule.max_minutes = parseInt(block.querySelector(".sched-rnd-max").value) || 120;
    }

    if (id) {
        config.push({
            id: id,
            enabled: isEnabled,
            image: imageName,
            phrases: phrases,
            require_confirmation: requireConfirm,
            snooze_minutes: snoozeMinutes,
            schedule: schedule
        });
    }
  });

  return config;
}

function sendProductivityAPIKey() {
  const isTodoist = getE("todoistEnable").checked;
  const provider = isTodoist ? "todoist" : "none";
  
  const formData = new FormData();
  formData.append("provider", provider);
  formData.append("target_robot", getE("targetBot").value);
  formData.append("key", getE("prodApiKey").value);
  
  const manualConfigArray = collectManualConfigData(formData);
  formData.append("manual_config", JSON.stringify(manualConfigArray));

  displayMessage("addProductivityProviderAPIStatus", "Saving...");

  fetch("/api/set_productivity_api", {
    method: "POST",
    body: formData
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addProductivityProviderAPIStatus", response);
    })
    .catch((error) => {
      displayMessage("addProductivityProviderAPIStatus", "Error saving settings: " + error);
    });
}

function updateProductivityAPI() {
  populateRobotList().then(() => {
      fetch("/api/get_productivity_api")
        .then((response) => response.json())
        .then((data) => {
          if (data) {
              getE("todoistEnable").checked = (data.provider === "todoist");
              getE("prodApiKey").value = data.key || "";
              
              if (data.target_robot) {
                  getE("targetBot").value = data.target_robot;
              }

              if (data.manual_config && data.manual_config.length > 2) { 
                  getE("enableManualReminders").checked = true;
                  toggleManualReminders();
                  try {
                      const config = JSON.parse(data.manual_config);
                      getE("manualRemindersContainer").innerHTML = ""; 
                      if (Array.isArray(config)) {
                          config.forEach(item => addReminderBlock(item));
                      }
                  } catch (e) {
                      console.error("Error parsing manual config", e);
                  }
              } else {
                 getE("enableManualReminders").checked = false;
                 toggleManualReminders();
              }

              checkProductivity();
          }
        })
        .catch(() => {
            checkProductivity();
        });
  });
}

function checkKG() {
  const provider = getE("kgProvider").value;
  const elements = [
    "houndifyInput",
    "togetherInput",
    "customAIInput",
    "intentGraphInput",
    "openAIInput",
    "saveChatInput",
    "llmCommandInput",
    "openAIVoiceForEnglishInput",
  ];

  elements.forEach((el) => (getE(el).style.display = "none"));

  if (provider) {
    if (provider === "houndify") {
      getE("houndifyInput").style.display = "block";
      getE("intentGraphInput").style.display = "block";
    } else if (provider === "openai") {
      getE("intentGraphInput").style.display = "block";
      getE("openAIInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
      getE("openAIVoiceForEnglishInput").style.display = "block";
    } else if (provider === "together") {
      getE("intentGraphInput").style.display = "block";
      getE("togetherInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
    } else if (provider === "custom") {
      getE("intentGraphInput").style.display = "block";
      getE("customAIInput").style.display = "block";
      getE("saveChatInput").style.display = "block";
      getE("llmCommandInput").style.display = "block";
    }
  }
}

function sendKGAPIKey() {
  const provider = getE("kgProvider").value;
  const data = {
    enable: true,
    provider,
    key: "",
    model: "",
    id: "",
    intentgraph: false,
    robotName: "",
    openai_prompt: "",
    openai_voice: "",
    openai_voice_with_english: false,
    save_chat: false,
    commands_enable: false,
    endpoint: "",
  };
  if (provider === "openai") {
    data.key = getE("openaiKey").value;
    data.openai_prompt = getE("openAIPrompt").value;
    data.intentgraph = getE("intentyes").checked
    data.save_chat = getE("saveChatYes").checked
    data.commands_enable = getE("commandYes").checked
    data.openai_voice = getE("openaiVoice").value
    data.openai_voice_with_english = getE("voiceEnglishYes").checked
  } else if (provider === "custom") {
    data.key = getE("customKey").value;
    data.model = getE("customModel").value;
    data.openai_prompt = getE("customAIPrompt").value;
    data.endpoint = getE("customAIEndpoint").value;
    data.intentgraph = getE("intentyes").checked
    data.save_chat = getE("saveChatYes").checked
    data.commands_enable = getE("commandYes").checked
  } else if (provider === "together") {
    data.key = getE("togetherKey").value;
    data.model = getE("togetherModel").value;
    data.openai_prompt = getE("togetherAIPrompt").value;
    data.intentgraph = getE("intentyes").checked;
    data.save_chat = getE("saveChatYes").checked
    data.commands_enable = getE("commandYes").checked
  } else if (provider === "houndify") {
    data.key = getE("houndKey").value;
    data.id = getE("houndID").value;
    data.intentgraph = getE("intentyes").checked
  } else {
    data.enable = false;
  }

  fetch("/api/set_kg_api", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      displayMessage("addKGProviderAPIStatus", response);
      alert(response);
    });
}

function deleteSavedChats() {
  if (confirm("Are you sure? This will delete all saved chats.")) {
    fetch("/api/delete_chats")
      .then((response) => response.text())
      .then(() => {
        alert("Successfully deleted all saved chats.");
      });
  }
}

function updateKGAPI() {
  fetch("/api/get_kg_api")
    .then((response) => response.json())
    .then((data) => {
      getE("kgProvider").value = data.provider;
      if (data.provider === "openai") {
        getE("openaiKey").value = data.key;
        getE("openAIPrompt").value = data.openai_prompt;
        getE("openaiVoice").value = data.openai_voice;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
        getE("voiceEnglishYes").checked = data.openai_voice_with_english
      } else if (data.provider === "together") {
        getE("togetherKey").value = data.key;
        getE("togetherModel").value = data.model;
        getE("togetherAIPrompt").value = data.openai_prompt;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
      } else if (data.provider === "custom") {
        getE("customKey").value = data.key;
        getE("customModel").value = data.model;
        getE("customAIPrompt").value = data.openai_prompt;
        getE("customAIEndpoint").value = data.endpoint;
        getE("commandYes").checked = data.commands_enable
        getE("intentyes").checked = data.intentgraph
        getE("saveChatYes").checked = data.save_chat
      } else if (data.provider === "houndify") {
        getE("houndKey").value = data.key;
        getE("houndID").value = data.id;
        getE("intentyes").checked = data.intentgraph
      }
      checkKG();
    });
}

function setSTTLanguage() {
  const data = { language: getE("languageSelection").value };

  displayMessage("languageStatus", "Setting...");

  fetch("/api/set_stt_info", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("downloading")) {
        displayMessage("languageStatus", "Downloading model...");
        updateSTTLanguageDownload();
      } else {
        displayMessage("languageStatus", response);
        getE("languageSelectionDiv").style.display = response.includes("success") ? "block" : "none";
      }
    });
}

function updateSTTLanguageDownload() {

  const interval = setInterval(() => {
    fetch("/api/get_download_status")
      .then((response) => response.text())
      .then((response) => {
        displayMessage("languageStatus", response.includes("not downloading") ? "Initiating download..." : response)
        if (response.includes("success") || response.includes("error")) {
          displayMessage("languageStatus", response);
          getE("languageSelectionDiv").style.display = "block";
          clearInterval(interval);
        }
      });
  }, 500);
}

function sendRestart() {
  fetch("/api/reset")
    .then((response) => response.text())
    .then((response) => {
      displayMessage("restartStatus", response);
    });
}

function hideEditIntents() {
  getE("editIntentForm").style.display = "none";
  getE("editIntentStatus").innerHTML = "";
}

function showEditIntents() {
  getE("editIntentForm").style.display = "block";
}

function displayMessage(elementId, message) {
  const element = getE(elementId);
  element.innerHTML = "";
  const p = document.createElement("p");
  p.textContent = message;
  element.appendChild(p);
}

function displayError(elementId, message) {
  const element = getE(elementId);
  element.innerHTML = "";
  const error = document.createElement("p");
  error.innerHTML = message;
  element.appendChild(error);
}

function toggleSection(sectionToToggle, sectionToClose, foldableID) {
  const toggleSect = getE(sectionToToggle);
  const closeSect = getE(sectionToClose);

  if (toggleSect.style.display === "block") {
    closeSection(toggleSect, foldableID);
  } else {
    openSection(toggleSect, foldableID);
    closeSection(closeSect, foldableID);
  }
}

function openSection(sectionID) {
  sectionID.style.display = "block";
}

function closeSection(sectionID) {
  sectionID.style.display = "none";
}

function updateColor(id) {
  const l_id = id.replace("section", "icon");
  const elements = document.getElementsByName("icon");

  elements.forEach((element) => {
    element.classList.remove("selectedicon");
    element.classList.add("nowselectedicon");
  });

  const targetElement = document.getElementById(l_id);
  targetElement.classList.remove("notselectedicon");
  targetElement.classList.add("selectedicon");
}


function showLog() {
  toggleVisibility(["section-intents", "section-log", "section-botauth", "section-version", "section-uicustomizer"], "section-log", "icon-Logs");
  logDivArea = getE("botTranscriptedTextArea");
  getE("logscrollbottom").checked = true;
  logP = document.createElement("p");
  GetLog = true
  const interval = setInterval(() => {
    if (!GetLog) {
      clearInterval(interval);
      return;
    }
    const url = getE("logdebug").checked ? "/api/get_debug_logs" : "/api/get_logs";
    fetch(url)
      .then((response) => response.text())
      .then((logs) => {
        logDivArea.innerHTML = logs || "No logs yet, you must say a command to Vector. (this updates automatically)";
        if (getE("logscrollbottom").checked) {
          logDivArea.scrollTop = logDivArea.scrollHeight;
        }
      });
  }, 500);
}

function checkUpdate() {
  displayMessage("cVersion", "Checking for updates...");
  displayMessage("aUpdate", "");
  displayMessage("cCommit", "");
  fetch("/api/get_version_info")
    // type VersionInfo struct {
    // 	FromSource      bool   `json:"fromsource"`
    // 	InstalledVer    string `json:"installedversion"`
    // 	InstalledCommit string `json:"installedcommit"`
    // 	CurrentVer      string `json:"currentver"`
    // 	CurrentCommit   string `json:"currentcommit"`
    // 	UpdateAvailable bool   `json:"avail"`
    // }
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("error")) {
        // <p id="cVersion"></p>
        // <p style="display: none;" id="cCommit"></p>
        // <p id="aUpdate"></p>
        displayMessage(
          "cVersion",
          "There was an error: " + response
        );
        getE("updateGuideLink").style.display = "none";
      } else {
        const parsed = JSON.parse(response);
        if (parsed.fromsource) {
          if (!parsed.avail) {
            displayMessage("aUpdate", `You are on the latest version.`);
            getE("updateGuideLink").style.display = "none";
          } else {
            displayMessage("aUpdate", `A newer version of WirePod (commit: ${parsed.currentcommit}) is available! Use this guide to update WirePod: `);
            getE("updateGuideLink").style.display = "block";
          }
          displayMessage("cVersion", `Installed Commit: ${parsed.installedcommit}`);
        } else {
          displayMessage("cVersion", `Installed Version: ${parsed.installedversion}`);
          displayMessage("cCommit", `Based on wire-pod commit: ${parsed.installedcommit}`);
          getE("cCommit").style.display = "block";
          if (parsed.avail) {
            displayMessage("aUpdate", `A newer version of WirePod (${parsed.currentversion}) is available! Use this guide to update WirePod: `);
            getE("updateGuideLink").style.display = "block";
          } else {
            displayMessage("aUpdate", "You are on the latest version.");
            getE("updateGuideLink").style.display = "none";
          }
        }
      }
    });
}

function showLanguage() {
  toggleVisibility(["section-weather", "section-restart", "section-kg", "section-productivity", "section-language"], "section-language", "icon-Language");
  fetch("/api/get_stt_info")
    .then((response) => response.json())
    .then((parsed) => {
      if (parsed.provider !== "vosk" && parsed.provider !== "whisper.cpp") {
        displayError("languageStatus", `To set the STT language, the provider must be Vosk or Whisper. The current one is '${parsed.sttProvider}'.`);
        getE("languageSelectionDiv").style.display = "none";
      } else {
        getE("languageSelectionDiv").style.display = "block";
        getE("languageSelection").value = parsed.language;
      }
    });
}

function showVersion() {
  toggleVisibility(["section-log", "section-botauth", "section-intents", "section-version", "section-uicustomizer"], "section-version", "icon-Version");
  checkUpdate();
}

function showIntents() {
  toggleVisibility(["section-log", "section-botauth", "section-intents", "section-version", "section-uicustomizer"], "section-intents", "icon-Intents");
}

function showWeather() {
  toggleVisibility(["section-weather", "section-restart", "section-language", "section-productivity", "section-kg"], "section-weather", "icon-Weather");
}

function showProductivity() {
  toggleVisibility(["section-weather", "section-restart", "section-language", "section-kg", "section-productivity"], "section-productivity", "icon-Productivity");
}

function showKG() {
  toggleVisibility(["section-weather", "section-restart", "section-language", "section-productivity", "section-kg"], "section-kg", "icon-KG");
}

function toggleVisibility(sections, sectionToShow, iconId) {
  if (sectionToShow != "section-log") {
    GetLog = false;
  }
  sections.forEach((section) => {
    getE(section).style.display = "none";
  });
  getE(sectionToShow).style.display = "block";
  updateColor(iconId);
}