var client = new HttpClient();

function refreshFaceList() {
var x = document.getElementById("faceList");
x.innerHTML = ""
fetch("/api-sdk/get_faces")
.then(response => response.text())
.then ((response) => {
  if (response.includes("null")) {
    alert("Error, no faces exist.")
  } else {
  jsonResp = JSON.parse(response)
  for (var i = 0; i < jsonResp.length; i++){
    var option = document.createElement("option");
    option.text = jsonResp[i]["name"]
    option.value = jsonResp[i]["face_id"] + ":" + jsonResp[i]["name"]
    x.add(option);
  }
}
})
}

refreshFaceList()

function renameFace() {
  var x = document.getElementById("faceList");
  oldFaceName = x.value.split(":")[1]
  faceId = x.value.split(":")[0]
  newFaceName = window.prompt('Enter the new name here:');
  fetch("/api-sdk/rename_face?oldname=" + oldFaceName + "&id=" + faceId + "&newname=" + newFaceName)
    .then (function(){alert("Success!"); refreshFaceList()})
}

function deleteFace() {
  var x = document.getElementById("faceList");
  faceId = x.value.split(":")[0]
  fetch("/api-sdk/delete_face?id=" + faceId)
    .then (function(){alert("Success!"); refreshFaceList()})
}
