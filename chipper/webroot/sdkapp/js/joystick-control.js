var urlParams = new URLSearchParams(window.location.search);
var esn = urlParams.get("serial");

// Movement state tracking
var isMoving = false;
var currentMovement = { lw: 0, rw: 0 };
var movementInterval = null;

// Head and lift state tracking
var headMoving = false;
var liftMoving = false;
var headDirection = null;
var liftDirection = null;
var headInterval = null;
var liftInterval = null;

// Speed control
var speedMultiplier = 0.5; // 50% default speed

// Joystick variables
var movementJoystick = null;
var movementKnob = null;
var isDragging = false;
var joystickCenter = { x: 0, y: 0 };
var maxDistance = 70; // Maximum distance from center

// when page is closed
window.onbeforeunload = function() {
  stopCamStream();
  sendForm("/api-sdk/mirror_mode?enable=false");
  sendForm("/api-sdk/release_behavior_control");
  stopAllMovement();
};

function initJoystickControl() {
  updateControlButtons();
  setupJoystick();
}

function setupJoystick() {
  movementJoystick = document.getElementById('movementJoystick');
  movementKnob = document.getElementById('movementKnob');

  if (!movementJoystick || !movementKnob) return;

  const rect = movementJoystick.getBoundingClientRect();
  joystickCenter.x = rect.width / 2;
  joystickCenter.y = rect.height / 2;

  // Mouse events
  movementKnob.addEventListener('mousedown', startDrag);
  document.addEventListener('mousemove', drag);
  document.addEventListener('mouseup', stopDrag);

  // Touch events for mobile
  movementKnob.addEventListener('touchstart', startDragTouch, { passive: false });
  document.addEventListener('touchmove', dragTouch, { passive: false });
  document.addEventListener('touchend', stopDragTouch);

  // Prevent context menu on mobile
  movementJoystick.addEventListener('contextmenu', e => e.preventDefault());
}

function startDrag(e) {
  e.preventDefault();
  isDragging = true;
}

function startDragTouch(e) {
  e.preventDefault();
  isDragging = true;
}

function drag(e) {
  if (!isDragging) return;
  e.preventDefault();

  const rect = movementJoystick.getBoundingClientRect();
  const centerX = rect.left + rect.width / 2;
  const centerY = rect.top + rect.height / 2;

  const deltaX = e.clientX - centerX;
  const deltaY = e.clientY - centerY;

  updateJoystickPosition(deltaX, deltaY);
}

function dragTouch(e) {
  if (!isDragging) return;
  e.preventDefault();

  const touch = e.touches[0];
  const rect = movementJoystick.getBoundingClientRect();
  const centerX = rect.left + rect.width / 2;
  const centerY = rect.top + rect.height / 2;

  const deltaX = touch.clientX - centerX;
  const deltaY = touch.clientY - centerY;

  updateJoystickPosition(deltaX, deltaY);
}

function updateJoystickPosition(deltaX, deltaY) {
  const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);

  if (distance > maxDistance) {
    const scale = maxDistance / distance;
    deltaX *= scale;
    deltaY *= scale;
  }

  // Update knob position
  movementKnob.style.transform = `translate(${-50 + deltaX}%, ${-50 + deltaY}%)`;

  // Calculate movement values
  calculateMovement(deltaX, deltaY);
}

function calculateMovement(deltaX, deltaY) {
  // Normalize values to -1 to 1 range
  const normalizedX = deltaX / maxDistance;
  const normalizedY = -deltaY / maxDistance; // Invert Y for intuitive movement

  // Calculate left and right wheel speeds
  // Forward/backward is the Y component
  // Turning is the X component

  const forward = normalizedY;
  const turn = normalizedX;

  // Calculate differential drive values
  let leftWheel = forward + turn;
  let rightWheel = forward - turn;

  // Clamp values to -1 to 1
  leftWheel = Math.max(-1, Math.min(1, leftWheel));
  rightWheel = Math.max(-1, Math.min(1, rightWheel));

  // Apply speed multiplier and scale to motor range (roughly -200 to 200)
  const maxSpeed = 200;
  const leftSpeed = Math.round(leftWheel * maxSpeed * speedMultiplier);
  const rightSpeed = Math.round(rightWheel * maxSpeed * speedMultiplier);

  // Only send commands if movement has changed significantly
  if (Math.abs(leftSpeed - currentMovement.lw) > 5 || Math.abs(rightSpeed - currentMovement.rw) > 5) {
    currentMovement.lw = leftSpeed;
    currentMovement.rw = rightSpeed;

    if (Math.abs(leftSpeed) < 10 && Math.abs(rightSpeed) < 10) {
      // Stop movement
      stopMovement();
    } else {
      // Send movement command
      sendMovementCommand(leftSpeed, rightSpeed);
    }
  }
}

function sendMovementCommand(leftSpeed, rightSpeed) {
  if (!isMoving) {
    isMoving = true;
  }
  sendForm(`/api-sdk/move_wheels?lw=${leftSpeed}&rw=${rightSpeed}`);
}

function stopMovement() {
  if (isMoving) {
    isMoving = false;
    sendForm("/api-sdk/move_wheels?lw=0&rw=0");
    currentMovement = { lw: 0, rw: 0 };
  }
}

function stopDrag(e) {
  if (!isDragging) return;
  isDragging = false;

  // Return knob to center
  movementKnob.style.transform = 'translate(-50%, -50%)';

  // Stop movement
  stopMovement();
}

function stopDragTouch(e) {
  if (!isDragging) return;
  isDragging = false;

  // Return knob to center
  movementKnob.style.transform = 'translate(-50%, -50%)';

  // Stop movement
  stopMovement();
}

function updateSpeed() {
  const slider = document.getElementById('speedSlider');
  const speedValue = document.getElementById('speedValue');
  speedMultiplier = slider.value / 100;
  speedValue.textContent = slider.value;
}

// Head movement functions
function startHeadMovement(direction) {
  if (headMoving) return;

  headMoving = true;
  headDirection = direction;

  const speed = direction === 'up' ? 2 : -2;
  sendForm(`/api-sdk/move_head?speed=${speed}`);

  // Continue sending commands while button is held
  headInterval = setInterval(() => {
    if (headMoving) {
      sendForm(`/api-sdk/move_head?speed=${speed}`);
    }
  }, 100);
}

function stopHeadMovement() {
  if (!headMoving) return;

  headMoving = false;
  headDirection = null;

  if (headInterval) {
    clearInterval(headInterval);
    headInterval = null;
  }

  sendForm("/api-sdk/move_head?speed=0");
}

// Lift movement functions
function startLiftMovement(direction) {
  if (liftMoving) return;

  liftMoving = true;
  liftDirection = direction;

  const speed = direction === 'up' ? 2 : -2;
  sendForm(`/api-sdk/move_lift?speed=${speed}`);

  // Continue sending commands while button is held
  liftInterval = setInterval(() => {
    if (liftMoving) {
      sendForm(`/api-sdk/move_lift?speed=${speed}`);
    }
  }, 100);
}

function stopLiftMovement() {
  if (!liftMoving) return;

  liftMoving = false;
  liftDirection = null;

  if (liftInterval) {
    clearInterval(liftInterval);
    liftInterval = null;
  }

  sendForm("/api-sdk/move_lift?speed=0");
}

function stopAllMovement() {
  stopMovement();
  stopHeadMovement();
  stopLiftMovement();
}

// switches items to disabled or enables them
function updateControlButtons() {
  const assumeControl = document.getElementById("assume_control");
  const releaseControl = document.getElementById("release_control");
  const mirrorOn = document.getElementById("mirror_on");
  const mirrorOff = document.getElementById("mirror_off");
  const textSay = document.getElementById("textSay");
  const controlButtons = document.querySelectorAll('.control-button');

  if (assumeControl.checked) {
    mirrorOn.disabled = false;
    mirrorOff.disabled = false;
    textSay.disabled = false;
    controlButtons.forEach(btn => btn.disabled = false);
  } else if (releaseControl.checked) {
    sendForm("/api-sdk/mirror_mode?enable=false");
    mirrorOff.checked = true;
    mirrorOn.disabled = true;
    mirrorOff.disabled = true;
    textSay.disabled = true;
    controlButtons.forEach(btn => btn.disabled = true);
    stopAllMovement();
  }
}

function goBackToSettings() {
  sendForm("/api-sdk/release_behavior_control");
  sendForm("/api-sdk/stop_cam_stream");
  stopAllMovement();
  window.location.href = "./settings.html?serial=" + esn;
}

function sendForm(formURL) {
  let xhr = new XMLHttpRequest();
  if (formURL.includes("?")) {
    formURL = formURL + "&serial=" + esn;
  } else {
    formURL = formURL + "?serial=" + esn;
  }
  xhr.open("POST", formURL);
  xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  xhr.send();
}

var stream = document.createElement("img");

function showCamStream() {
  stream.src = "/cam-stream?serial=" + esn;
  document.getElementById("camStream").appendChild(stream);
}

function stopCamStream() {
  stream.src = "";
  if (stream.parentNode) {
    document.getElementById("camStream").removeChild(stream);
  }
  sendForm("/api-sdk/stop_cam_stream");
}

function sayText() {
  const sayTextValue = document.getElementById("textSay").value;
  sendForm("/api-sdk/say_text?text=" + encodeURIComponent(sayTextValue));
}

// Prevent accidental navigation away from page
window.addEventListener('beforeunload', function(e) {
  stopAllMovement();
});

// Handle page visibility changes (e.g., tab switching)
document.addEventListener('visibilitychange', function() {
  if (document.hidden) {
    stopAllMovement();
  }
});
