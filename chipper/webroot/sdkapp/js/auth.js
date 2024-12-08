var client = new HttpClient();

var botList = document.getElementById("botList");

getSDKInfo().then((jsonResp) => {
  if (!botList) {
    return;
  }
  for (var i = 0; i < jsonResp["robots"].length; i++) {
    var option = document.createElement("option");
    option.text = jsonResp["robots"][i]["esn"];
    option.value = jsonResp["robots"][i]["esn"];
    botList.add(option);
  }
}).catch((error) => {
  console.error('Unable to get SDK info:', error);
  alert("Error getting robot list. This either means that no robots are authenticated or a previously authenticated robot is not connected.");
  window.location.href = "/";
});

function connectSDK() {
  var botList = document.getElementById("botList");
  fetch("/api-sdk/conn_test?serial=" + botList.value)
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        window.location.href = "./settings.html?serial=" + botList.value;
      } else {
        alert(response);
      }
    });
}
