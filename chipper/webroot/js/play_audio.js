
//Convert the WAV frame rate on the client machine and send the WAV file to be processed at /api-sdk/play_sound

let processedAudioBlob = null; 

document.getElementById('fileInput').addEventListener('change', async () => {
    const fileInput = document.getElementById('fileInput');
    if (!fileInput.files.length) {
        alert('Please, select a WAV file');
        return;
    }

    const file = fileInput.files[0];
    const arrayBuffer = await file.arrayBuffer();
    const audioContext = new (window.AudioContext || window.webkitAudioContext)();

    try {
        const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);

        let newBuffer;

        //Verify is audio is stereio or mono and convert to mono if necessary
        if(audioBuffer.numberOfChannels >1){
            const monoLength = audioBuffer.length;
            newBuffer = audioContext.createBuffer(1, monoLength, audioBuffer.sampleRate);

            const channelData = newBuffer.getChannelData(0);
            for (let i = 0; i < monoLength; i++){

                channelData[i] = 0.5 * (audioBuffer.getChannelData(0)[i] + audioBuffer.getChannelData(1)[i]);

                
            }

        }else{
            // if already mono, just copy
            newBuffer = audioBuffer;
            }
            
        

        // adjust frame rate to 8000 Hz
        const newSampleRate = 8000;
        const newLength = Math.round(newBuffer.length * newSampleRate / newBuffer.sampleRate);
        const resampledBuffer  = audioContext.createBuffer(newBuffer.numberOfChannels, newLength, newSampleRate);

        for (let channel = 0; channel < newBuffer.numberOfChannels; channel++) {
            const oldData = newBuffer.getChannelData(channel);
            const newData = resampledBuffer .getChannelData(channel);

            for (let i = 0; i < newLength; i++) {
                const oldIndex = i * newBuffer.sampleRate / newSampleRate;
                const index0 = Math.floor(oldIndex);
                const index1 = Math.min(index0 + 1, oldData.length - 1);
                const fraction = oldIndex - index0;

                newData[i] = oldData[index0] * (1 - fraction) + oldData[index1] * fraction; 
            }
        }

        // Create a new WAV file
        processedAudioBlob = await bufferToWave(resampledBuffer );
        const url = URL.createObjectURL(processedAudioBlob);

        // play processed audio
        const audioOutput = document.getElementById('audioOutput');
        audioOutput.src = url;
        audioOutput.play();

        // show send button
        document.getElementById('uploadButton').style.display = 'inline-block';
    } catch (error) {
        console.error('Error processing the file:', error);
    }
});

function bufferToWave(abuffer) {
    const numOfChannels = abuffer.numberOfChannels;
    const length = abuffer.length * numOfChannels * 2 + 44;
    const buffer = new ArrayBuffer(length);
    const view = new DataView(buffer);
    let offset = 0;

    // Escrever cabeçalho WAV
    setString(view, offset, 'RIFF'); offset += 4;
    view.setUint32(offset, length - 8, true); offset += 4;
    setString(view, offset, 'WAVE'); offset += 4;
    setString(view, offset, 'fmt '); offset += 4;
    view.setUint32(offset, 16, true); offset += 4; // format size
    view.setUint16(offset, 1, true); offset += 2; // PCM format
    view.setUint16(offset, numOfChannels, true); offset += 2; // number of channels
    view.setUint32(offset, 8000, true); offset += 4; // sample rate
    view.setUint32(offset, 8000 * numOfChannels * 2, true); offset += 4; // byte rate
    view.setUint16(offset, numOfChannels * 2, true); offset += 2; // block align
    view.setUint16(offset, 16, true); offset += 2; // bits per sample

    setString(view, offset, 'data'); offset += 4;
    view.setUint32(offset, length - offset - 4, true); offset += 4;

    // Copiar os dados de áudio
    for (let channel = 0; channel < numOfChannels; channel++) {
        const channelData = abuffer.getChannelData(channel);
        for (let i = 0; i < channelData.length; i++) {
            view.setInt16(offset, channelData[i] * 0x7FFF, true);
            offset += 2;
        }
    }

    return new Blob([buffer], { type: 'audio/wav' });
}

function setString(view, offset, string) {
    for (let i = 0; i < string.length; i++) {
        view.setUint8(offset + i, string.charCodeAt(i));
    }
}

document.getElementById('uploadButton').addEventListener('click', async () => {
    if (!processedAudioBlob) {
        alert('No processed audio to send.');
        return;
    }

    await uploadAudio(processedAudioBlob);
});

async function uploadAudio(blob) {
    const formData = new FormData();
    formData.append('sound', blob, 'processed.wav');

    try {
        const dominio = window.location.hostname;
        esn = urlParams.get("serial");
        const response = await fetch('/api-sdk/play_sound?serial='+esn, {
            method: 'POST',
            body: formData,
        });

        if (!response.ok) {
            throw new Error('Erro to send audio: ' + response.statusText);
        }

        const result = await response.json();
        console.log('Audio sent successfully:', result);
    } catch (error) {
        console.error('Error sending audio:', error);
    }
}
