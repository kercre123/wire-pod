function img_recognize() {
    response = sendForm("/api-sdk/image_detection");
    if (response) {
        // Aqui você pode fazer algo com a resposta
        console.log("Resposta do servidor:", response);
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

    xhr.onload = function () {
        if (xhr.status === 200) {
            // Captura a string base64 retornada pelo servidor
            let base64String = xhr.responseText;
            console.log("Base64 String:", base64String);
            // Aqui você pode fazer algo com a string base64
        } else {
            console.error("Erro na requisição:", xhr.status, xhr.statusText);
        }
    };

    xhr.onerror = function () {
        console.error("Erro de rede ou CORS.");
    };

    xhr.send();
    return xhr.responseText;

}