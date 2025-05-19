async function img_recognize() {
    //assume behavior control
    sendForm('/api-sdk/assume_behavior_control?priority=high')
    //send image detection request
    try {
        let response = await sendForm("/api-sdk/image_detection");
        console.log("Teste: " + response);
        sendForm("/api-sdk/say_text?text=I see " + response);
    } catch (error) {
        console.error('Error:', error);
    }
    //release behavior control
    await new Promise(resolve => setTimeout(resolve, 2000));
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
