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
    x.add(option);
  }
}).catch(() => {
  alert("Error, it's likely no bots are authenticated");
});

function connectSDK() {
  var x = document.getElementById("botList");
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
