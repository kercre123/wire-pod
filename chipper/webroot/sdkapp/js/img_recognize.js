function img_recognize() {
    sendForm("/api-sdk/image_detection");

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