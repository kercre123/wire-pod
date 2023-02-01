var colorPicker = new iro.ColorPicker("#picker", {
  width: 250,
  layout: [
  { 
    component: iro.ui.Wheel,
  }
  ]
});

var stimRunning = false

var urlParams = new URLSearchParams(window.location.search);
esn = urlParams.get('serial');

var client = new HttpClient();
getCurrentSettings()

function revealSdkActions() {
  var x = document.getElementById("sdkActions");
  x.style.display = "block";
}

function showSection(id) {
  var headings = document.getElementsByClassName("toggleable-section");
  for (var i = 0; i < headings.length; i++) {
      headings[i].style.display = "none";
  }
  document.getElementById(id).style.display = "block";
  updateColor(id);
  if (id == "section-stim") {
    logDiv = document.getElementById("stimStatus")
    logP = document.createElement("p")
    stimRunning = true
    sendForm("/api-sdk/begin_event_stream")
    interval = setInterval(function() {
      if (stimRunning == false) {
          sendForm("/api-sdk/stop_event_stream")
          clearInterval(interval)
          return
      }
      let xhr = new XMLHttpRequest();
      xhr.open("GET", "/api-sdk/get_stim_status?serial=" + esn);
      xhr.send();
      xhr.onload = function() {
              logDiv.innerHTML = ""
              logP.innerHTML = xhr.response + "/1.00000000"
              logDiv.appendChild(logP)
      }
  }, 1000)
  
  } else {
    stimRunning = false
  }
}

function sendForm(formURL) {
  let xhr = new XMLHttpRequest();
    if (formURL.includes("?")) {
      formURL = formURL + "&serial=" + esn
    } else {
      formURL = formURL + "?serial=" + esn
    }
    xhr.open("POST", formURL);
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
    xhr.send();
    xhr.onload = function() { 
      getCurrentSettings()
    };
}

function goToControlPage() {
    window.location.href = './control.html?serial=' + esn
}

function sendLocation() {
  locationInput = document.getElementById("locationInput").value
  console.log(locationInput)
  if (locationInput == "") {
    alert("Location cannot be blank.")
    return
  }
  let xhr = new XMLHttpRequest();
    xhr.open("POST", "/api-sdk/location?serial=" + esn + "&location=" + locationInput);
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
    xhr.send();
    xhr.onload = function() { 
      getCurrentSettings()
    };
}

function sendTimeZone() {
  timezone = document.getElementById("tzInput").value
  if (timezone == "") {
    alert("Time zone cannot be blank.")
    return
  }
  let xhr = new XMLHttpRequest();
    xhr.open("POST", "/api-sdk/timezone?serial=" + esn + "&timezone=" + timezone);
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
    xhr.send();
    xhr.onload = function() { 
      getCurrentSettings()
    };
}

function sendCustomColor() {
  var pickerHue = colorPicker.color.hue;
  var pickerSat = colorPicker.color.saturation;
  var sendHue = pickerHue / 360
  var sendHue = sendHue.toFixed(3)
  var sendSat = pickerSat / 100
  var sendSat = sendSat.toFixed(3)
  let data = "hue=" + sendHue + "&" + "sat=" + sendSat
  let xhr = new XMLHttpRequest();
  xhr.open("POST", "/api-sdk/custom_eye_color?serial=" + esn);
  xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  xhr.send(data);
  xhr.onload = function() { 
    getCurrentSettings()
  };
};

function getCurrentSettings() {
  let xhr = new XMLHttpRequest();
  xhr.open("POST", "/api-sdk/get_sdk_settings?serial=" + esn);
  xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
  xhr.responseType = 'json';
  xhr.send();
  xhr.onload = function() {
    var jdocSdkSettingsResponse1 = JSON.stringify(xhr.response)
    jdocSdk = JSON.parse(jdocSdkSettingsResponse1)
    let xhr2 = new XMLHttpRequest();
      if ( jdocSdk["custom_eye_color"]) {
        var customECE = jdocSdk["custom_eye_color"]["enabled"]
        var customECH = jdocSdk["custom_eye_color"]["hue"]
        var customECS = jdocSdk["custom_eye_color"]["saturation"]
      }
      eyeColorS = jdocSdk["eye_color"]
      var volumeS = jdocSdk["master_volume"]
      var localeS = jdocSdk["locale"]
      var timeSetS = jdocSdk["clock_24_hour"]
      var tempFormatS = jdocSdk["temp_is_fahrenheit"]
      var buttonS = jdocSdk["button_wakeword"]
      var location = jdocSdk["default_location"]
      var timezone = jdocSdk["time_zone"]
      if ( jdocSdk["custom_eye_color"]) {
       if (`${customECE}` == "true") {
         var setHue = customECH * 360
         var setHue = setHue.toFixed(3)
         var setSat = customECS * 100
         var setSat = setSat.toFixed(3)
         colorPicker.color.hsl = { h: setHue, s: setSat, l: 50 };     
         var eyeColorT = "Custom"
       } else { 
        if (`${eyeColorS}` == 0) {
          eyeColorT = "Teal"
        } else if (`${eyeColorS}` == 1) {
          eyeColorT = "Orange"
        } else if (`${eyeColorS}` == 2) {
          eyeColorT = "Yellow"
        } else if (`${eyeColorS}` == 3) {
          eyeColorT = "Lime Green"
        } else if (`${eyeColorS}` == 4) {
          eyeColorT = "Azure Blue"
        } else if (`${eyeColorS}` == 5) {
          eyeColorT = "Purple"
        } else if (`${eyeColorS}` == 6) {
          eyeColorT = "Other Green"
        } else {
          eyeColorT = "none"
        }
      } } else { 
       if (`${eyeColorS}` == 0) {
        eyeColorT = "Teal"
      } else if (`${eyeColorS}` == 1) {
        eyeColorT = "Orange"
      } else if (`${eyeColorS}` == 2) {
        eyeColorT = "Yellow"
      } else if (`${eyeColorS}` == 3) {
        eyeColorT = "Lime Green"
      } else if (`${eyeColorS}` == 4) {
        eyeColorT = "Azure Blue"
      } else if (`${eyeColorS}` == 5) {
        eyeColorT = "Purple"
      } else if (`${eyeColorS}` == 6) {
        eyeColorT = "Other Green"
      } else {
        eyeColorT = "none"
      }
    }
    if (`${volumeS}` == 0) {
      var volumeT = "Mute"
    } else if (`${volumeS}` == 1) {
      var volumeT = "Low"
    } else if (`${volumeS}` == 2) {
      var volumeT = "Medium Low"
    } else if (`${volumeS}` == 3) {
      var volumeT = "Medium"
    } else if (`${volumeS}` == 4) {
      var volumeT = "Medium High"
    } else if (`${volumeS}` == 5) {
      var volumeT = "High"
    } else {
      var volumeT = "none"
    }
    if (`${timeSetS}` == "false") {
      var timeSetT = "12 Hour"
    } else {
      var timeSetT = "24 Hour"
    }
    if (`${tempFormatS}` == "true") {
      var tempFormatT = "Fahrenheit"
    } else {
      var tempFormatT = "Celcius"
    }
    if (`${buttonS}` == 0) {
      var buttonT = "Hey Vector"
    } else {
      var buttonT = "Alexa"
    }
    
    var s1 = document.getElementById('currentVolume');
    const s1P = document.createElement('p');
    document.getElementById(volumeT).checked = true;

    var s2 = document.getElementById('currentEyeColor');
    const s2P = document.createElement('p');
    if (eyeColorT != "none" && eyeColorT != "Custom") {
      document.getElementById(eyeColorT).checked = true;
    }
    
    var s3 = document.getElementById('currentLocale');
    const s3P = document.createElement('p');
    document.getElementById(localeS).checked = true;
    
    var s4 = document.getElementById('currentTimeSet');
    const s4P = document.createElement('p');
    document.getElementById(timeSetT).checked = true;
    
    var s5 = document.getElementById('currentTempFormat');
    const s5P = document.createElement('p');
    document.getElementById(tempFormatT).checked = true;
    
    var s6 = document.getElementById('currentButton');
    const s6P = document.createElement('p');
    document.getElementById(buttonT).checked = true;
    
    var s10 = document.getElementById('currentLocation');
    const s10P = document.createElement('p');
    s10P.textContent = "Current Location Setting: " + `${location}`
    document.getElementById('locationInput').placeholder = `${location}`
    s10.innerHTML = ''
    s10.appendChild(s10P);
    
    var s11 = document.getElementById('currentTimeZone');
    const s11P = document.createElement('p');
    s11P.textContent = "Current Time Zone Setting: " + `${timezone}`
    document.getElementById('tzInput').placeholder = `${timezone}`
    s11.innerHTML = ''
    s11.appendChild(s11P);
  };
};


