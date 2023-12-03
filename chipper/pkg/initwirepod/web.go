package initwirepod

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	botsetup "github.com/kercre123/wire-pod/chipper/pkg/wirepod/setup"
)

// cant be part of config-ws, otherwise import cycle

func ChipperHTTPApi(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api-chipper/restart":
		RestartServer()
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-chipper/use_ip":
		port := r.FormValue("port")
		if port == "" {
			fmt.Fprint(w, "error: must have port")
			return
		}
		_, err := strconv.Atoi(port)
		if err != nil {
			fmt.Fprint(w, "error: port is invalid")
			return
		}
		vars.APIConfig.Server.EPConfig = false
		vars.APIConfig.Server.Port = port
		err = botsetup.CreateCertCombo()
		botsetup.CreateServerConfig()
		if err != nil {
			logger.Println(err)
			fmt.Fprint(w, "error: "+err.Error())
			return
		}
		vars.APIConfig.PastInitialSetup = true
		vars.WriteConfigToDisk()
		RestartServer()
		fmt.Fprint(w, "done")
		return
	case r.URL.Path == "/api-chipper/use_ep":
		vars.APIConfig.Server.EPConfig = true
		vars.APIConfig.Server.Port = "443"
		vars.APIConfig.PastInitialSetup = true
		botsetup.CreateServerConfig()
		vars.WriteConfigToDisk()
		RestartServer()
		fmt.Fprint(w, "done")
		return
	}
}
