var client = new HttpClient();

async function getSDKInfo() {
  try {
    var response = await fetch("/api-sdk/get_sdk_info");
    if (!response.ok) {
      return undefined;
    }
    var data = await response.json();
    return data;
  } catch (error) {
    console.error('Unable to get SDK info:', error);
    throw error;
  }
}

var x = document.getElementById("botList");

getSDKInfo().then((jsonResp) => {
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
  fetch("/api-sdk/conn_test?serial=" + x.value)
    .then((response) => response.text())
    .then((response) => {
      if (response.includes("success")) {
        window.location.href = "./settings.html?serial=" + x.value;
      } else {
        alert(response);
      }
    });
}
