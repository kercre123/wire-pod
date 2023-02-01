var client = new HttpClient();

var x = document.getElementById("botList");
fetch("/api-sdk/get_sdk_info")
.then(response => response.text())
.then ((response) => {
  if (response.includes("error")) {
    alert("Error, it is likely no bots are authenticated.")
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
  fetch("/api-sdk/conn_test?serial=" + x.value)
  .then (response => response.text())
  .then ((response) => {
    if (response.includes("success")) {
      window.location.href = './settings.html?serial=' + x.value
    } else {
      alert(response)
    }
  })
}
