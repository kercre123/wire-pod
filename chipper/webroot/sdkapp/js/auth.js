var client = new HttpClient();
function certReset() {
  let confirmAction = confirm("This will put Vector back on the Onboarding screen and he will be unauthenticated from his account. You should use this page or the Vector mobile app to authenticate him with an Anki account after this process is complete. Vector's stats and personality will not be changed or erased. Would you like to continue?");
  if (confirmAction) {
    fetch("/api-sdk/server_prod")
    alert("Executing. Vector's eyes will disappear and his face will show 'configuring...'. After a while, he will boot back up to the onboarding screen (blinking V). Once he is there, press OK and try using this app to authenticate him.");
  }
}

var x = document.getElementById("botList");
fetch("/api-sdk/get_sdk_info")
.then(response => response.text())
.then ((response) => {
  if (response.includes("error")) {
    alert("Error, it is likely no bots are authenticated. Debug: " + response)
  } else {
  jsonResp = JSON.parse(response)
  for (var i = 0; i < jsonResp["robots"].length; i++){
    var option = document.createElement("option");
    option.text = jsonResp["robots"][i]["esn"]
    option.value = jsonResp["robots"][i]["esn"]
    x.add(option);
  }
}
})

function connectSDK() {
  var x = document.getElementById("botList");
  fetch("/api-sdk/initSDK?serial=" + x.value)
  .then (response => response.text())
  .then ((response) => {
    if (response.includes("success")) {
      window.location.href = './settings.html'
    } else {
      alert(response)
    }
  })
}

function submitCreds() {
  const form = document.getElementById('authForm');
  event.preventDefault();
  var usernameForm = form.elements['username'];
  var passwordForm = form.elements['password'];
  let emailSend = usernameForm.value;
  let passSend = passwordForm.value;
  var data = "username=" + emailSend + "&password=" + passSend
  var client = new HttpClient();
  var result = document.getElementById('authResult');
  const resultP = document.createElement('p');
  resultP.textContent =  "Authenticating...";
  result.innerHTML = '';
  result.appendChild(resultP);
  fetch("/api-sdk/sdk_auth?" + data)
  .then(response => response.text())
  .then((response) => {
    res = response.replace(/\s/g,'');
    result.innerHTML = '';
    if (`${res}` == "success") {
      resultP.textContent = "Authentication successful! Now you can use the app." 
      result.appendChild(resultP);
      const resultA = document.createElement('a');
      var resultAtext = document.createTextNode("Click here to return to the app")
      resultA.appendChild(resultAtext);
      resultA.title = "Click here to authorize";
      resultA.href = "/";
      result.appendChild(resultA);     
    } else if (`${res}` == "error") {
      resultP.textContent = "Invalid username or password. Please try again."
      result.appendChild(resultP);
    } else if (`${res}` == "error2") {
      resultP.textContent = "This account is valid, but the robot did not accept it. Try another Anki account. If nothing works, press the button below to put the bot back into onboarding mode so you can reauthenticate him with the Vector mobile app (or this page). This will not clear user data."
      result.appendChild(resultP);
    } else {
      resultP.textContent = "An unknown error has occurred."
      result.appendChild(resultP);
    };
  });
};
