async function img_recognize() {
    sendForm('/api-sdk/assume_behavior_control?priority=high')
    try {
    let response = await sendForm("/api-sdk/image_detection");
    console.log("Teste: "+response);
    sendForm("/api-sdk/say_text?text=" + response);
    } catch (error) {
        console.error('Error:', error);
    }
   sendForm("/api-sdk/release_behavior_control")
}

async function sendForm(formURL) {

    if (formURL.includes("?")) {
        formURL = formURL + "&serial=" + esn;
    } else {
        formURL = formURL + "?serial=" + esn;
    }
    let response = await fetch(formURL, {
        method: 'POST',
        headers: {
           'Content-Type': 'application/x-www-form-urlencoded',
        }
    });
    if (!response.ok) {
        throw new Error('Request Error: ' + response.statusText);
    }
    return await response.text();
}
