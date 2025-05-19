async function img_recognize() {
    try {
    let response = await sendForm("/api-sdk/image_detection");
    console.log("Teste: "+response);
    } catch (error) {
        console.error('Error:', error);
    }
   
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