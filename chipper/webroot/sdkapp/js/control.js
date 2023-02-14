var x = document.getElementById("sdkActions");
var keysKey = document.getElementById("keysKey");
var useKeyboardControl = false
var camStream = document.getElementById("camStream")
var urlParams = new URLSearchParams(window.location.search);
esn = urlParams.get('serial');


// wheels
var isMovingForward = false
var isMovingLeft = false
var isMovingRight = false
var isMovingFL = false
var isMovingFR = false
var isMovingBack = false
var isMovingBL = false
var isMovingBR = false
var isStopped = false

// lift
var liftIsMovingDown = false
var liftIsMovingUp = false
var liftIsStopped = false

// head
var headIsMovingDown = false
var headIsMovingUp = false
var headIsStopped = false

function toggleKeyboard() {
    if (useKeyboardControl == false) {
        useKeyboardControl = true
        keysKey.style.display = "block"
    } else {
        useKeyboardControl = false
        keysKey.style.display = "none"
    }
}

function goBackToSettings() {
    sendForm('/api-sdk/release_behavior_control')
    sendForm('/api-sdk/stop_cam_stream')
    window.location.href = './settings.html?serial=' + esn
}

function sdkUnInit() {
    sendForm('/api-sdk/release_behavior_control')
    sendForm('/api-sdk/stop_cam_stream')
    var x = document.getElementById("sdkActions");
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
  }

let keysPressed = {};

var stream = document.createElement("img");

function showCamStream() {
    //sendForm('/api-sdk/begin_cam_stream')
    stream.src = "/cam-stream?serial=" + esn;
    document.getElementById("camStream").appendChild(stream)
}

function stopCamStream() {
    stream.src = ""
    document.getElementById("camStream").removeChild(stream)
    sendForm('/api-sdk/stop_cam_stream')
}

function sayText() {
    sayTextValue = document.getElementById("textSay").value
    sendForm("/api-sdk/say_text?text=" + sayTextValue)
}

keysPressed["w"] = false
keysPressed["a"] = false
keysPressed["s"] = false
keysPressed["d"] = false

keysPressed["r"] = false
keysPressed["f"] = false

keysPressed["t"] = false
keysPressed["g"] = false


document.addEventListener('keyup', (event) => {
    keysPressed[event.key] = false
    sendControlRequests()
 });

document.addEventListener('keydown', function(event) {
    keysPressed[event.key] = true;
    sendControlRequests()
});

function sendControlRequests() {
    if (useKeyboardControl == true) {
        if (keysPressed["w"] == false && keysPressed["a"] == false && keysPressed["s"] == false && keysPressed["d"] == false) {
            if (isStopped == false) {
                sendForm("/api-sdk/move_wheels?lw=0&rw=0")
                sendForm("/api-sdk/move_wheels?lw=0&rw=0")
            }
            isStopped = true
            isMovingForward = false
            isMovingBL = false
            isMovingBR = false
            isMovingBack = false
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = false
        } else if (keysPressed["w"] == true && keysPressed["a"] == false && keysPressed["s"] == false && keysPressed["d"] == false) {
            if (isMovingForward == false) {
                sendForm("/api-sdk/move_wheels?lw=140&rw=140")
            }
            isStopped = false
            isMovingForward = true
            isMovingBL = false
            isMovingBR = false
            isMovingBack = false
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = false
        } else if (keysPressed["w"] == true && keysPressed["a"] == true && keysPressed["s"] == false && keysPressed["d"] == false) {
            if (isMovingFL == false) {
                sendForm("/api-sdk/move_wheels?lw=100&rw=190")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = false
            isMovingBR = false
            isMovingBack = false
            isMovingFL = true
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = false
        } else if (keysPressed["w"] == true && keysPressed["a"] == false && keysPressed["s"] == false && keysPressed["d"] == true) {
            if (isMovingFR == false) {
                sendForm("/api-sdk/move_wheels?lw=190&rw=100")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = false
            isMovingBR = false
            isMovingBack = false
            isMovingFL = false
            isMovingFR = true
            isMovingLeft = false
            isMovingRight = false
        } else if (keysPressed["w"] == false && keysPressed["a"] == false && keysPressed["s"] == false && keysPressed["d"] == true) {
            if (isMovingRight == false) {
                sendForm("/api-sdk/move_wheels?lw=150&rw=-150")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = false
            isMovingBR = false
            isMovingBack = false
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = true
        } else if (keysPressed["w"] == false && keysPressed["a"] == true && keysPressed["s"] == false && keysPressed["d"] == false) {
            if (isMovingLeft == false) {
                sendForm("/api-sdk/move_wheels?lw=-150&rw=150")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = false
            isMovingBR = false
            isMovingBack = false
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = true
            isMovingRight = false
        } else if (keysPressed["w"] == false && keysPressed["a"] == false && keysPressed["s"] == true && keysPressed["d"] == false) {
            if (isMovingBack == false) {
                sendForm("/api-sdk/move_wheels?lw=-150&rw=-150")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = false
            isMovingBR = false
            isMovingBack = true
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = false
        } else if (keysPressed["w"] == false && keysPressed["a"] == true && keysPressed["s"] == true && keysPressed["d"] == false) {
            if (isMovingBL == false) {
                sendForm("/api-sdk/move_wheels?lw=-100&rw=190")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = true
            isMovingBR = false
            isMovingBack = false
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = false
        } else if (keysPressed["w"] == false && keysPressed["a"] == false && keysPressed["s"] == true && keysPressed["d"] == true) {
            if (isMovingBR == false) {
                sendForm("/api-sdk/move_wheels?lw=-190&rw=100")
            }
            isStopped = false
            isMovingForward = false
            isMovingBL = false
            isMovingBR = true
            isMovingBack = false
            isMovingFL = false
            isMovingFR = false
            isMovingLeft = false
            isMovingRight = false
        }
        if (keysPressed["r"] == false && keysPressed["f"] == false) {
            if (liftIsStopped == false) {
                sendForm("/api-sdk/move_lift?speed=0")
            }
            liftIsStopped = true
            liftIsMovingDown = false
            liftIsMovingUp = false
        } else if (keysPressed["r"] == true && keysPressed["f"] == false) {
            if (liftIsMovingUp == false) {
                sendForm("/api-sdk/move_lift?speed=2")
            }
            liftIsStopped = false
            liftIsMovingDown = false
            liftIsMovingUp = true
        } else if (keysPressed["r"] == false && keysPressed["f"] == true) {
            if (liftIsMovingUp == false) {
                sendForm("/api-sdk/move_lift?speed=-2")
            }
            liftIsStopped = false
            liftIsMovingDown = true
            liftIsMovingUp = false
        }
        if (keysPressed["t"] == false && keysPressed["g"] == false) {
            if (headIsStopped == false) {
                sendForm("/api-sdk/move_head?speed=0")
            }
            headIsStopped = true
            headIsMovingDown = false
            headIsMovingUp = false
        } else if (keysPressed["t"] == true && keysPressed["g"] == false) {
            if (headIsMovingUp == false) {
                sendForm("/api-sdk/move_head?speed=2")
            }
            headIsStopped = false
            headIsMovingDown = false
            headIsMovingUp = true
        } else if (keysPressed["t"] == false && keysPressed["g"] == true) {
            if (headIsMovingUp == false) {
                sendForm("/api-sdk/move_head?speed=-2")
            }
            headIsStopped = false
            headIsMovingDown = true
            headIsMovingUp = false
        }
    }
}
