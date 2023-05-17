function updateSSHStatus(statusString) {
    setupStatus = document.getElementById("oskrSetupProgress")
    setupStatus.innerHTML = ""
    setupStatusP = document.createElement("p")
    setupStatusP.innerHTML = statusString
    setupStatus.appendChild(setupStatusP)
}

function doSSHSetup() {
  const ip = document.getElementById('sshIp').value;
  const key = document.getElementById('sshKeyFile').files[0];
  
  if (ip && key) {
    const formData = new FormData();
    formData.append('key', key);
    formData.append('ip', ip);
    
    fetch('/api-ssh/setup', {
      method: 'POST',
      body: formData
    })
    .then(response => response.text())
    .then((response => {
      if (response.includes("running")) {
        document.getElementById("oskrSetup").style.display = "none"
        updateSSHSetup()
        return
      } else {
        updateSSHStatus(response)
      }
    })
  )
} else {
  updateSSHStatus("You must enter an IP address and upload a key.")
}
}

function updateSSHSetup() {
    interval = setInterval(function(){
    fetch("/api-ssh/get_setup_status")
      .then(response => response.text())
      .then((response => {
        statusText = response
        if (response.includes("done")) {
          updateSSHStatus("File transfer complete! Use the above section to complete bot setup. The bot should eventually be on the onboarding screen.")
          document.getElementById("oskrSetup").style.display = "block"
          clearInterval(interval)
        } else if (response.includes("error")) {
          resp = response
          if (response.includes("no route to host")) {
            resp = "Wire-pod was unable to connect to the robot. Make sure the robot is running OSKR/dev software and that it is on the same network as this wire-pod instance. Also double-check the IP."
          }
          updateSSHStatus(resp)
            clearInterval(interval)
            document.getElementById("oskrSetup").style.display = "block"
            return
        } else if (response.includes("not running")) {
          updateSSHStatus("Initiating SSH transfer...")
        } else {
          updateSSHStatus(response)
        }
      }))
  }, 500)
}