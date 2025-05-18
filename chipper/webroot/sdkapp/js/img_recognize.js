function img_recognize() {
    request = sendForm("/api-sdk/api-sdk/image_detection");
    request.onreadystatechange = function() {
        if (request.readyState === 4 && request.status === 200) {
            const response = JSON.parse(request.responseText);
            if (response && response.base64String) {
            const imageUrl = "data:image/png;base64," + response.base64String;
            const newTab = window.open();
            if (newTab) {
                newTab.document.body.innerHTML = `<img src="${imageUrl}" alt="Recognized Image">`;
            } else {
                console.error("Failed to open a new tab");
            }
            } else {
            console.error("Base64 string not found in response");
            return null;
            }
        }
    }

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