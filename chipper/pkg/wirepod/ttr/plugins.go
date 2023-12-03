package wirepod_ttr

import (
	"os"
	"plugin"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var PluginList []*plugin.Plugin
var PluginUtterances []*[]string
var PluginFunctions []func(string, string) (string, string)
var PluginNames []string

func LoadPlugins() {
	logger.Println("Loading plugins")
	entries, err := os.ReadDir("./plugins")
	if err != nil {
		logger.Println("Unable to load plugins:")
		logger.Println(err)
	}
	for _, file := range entries {
		if strings.Contains(file.Name(), ".so") {
			plugin, err := plugin.Open("./plugins/" + file.Name())
			if err != nil {
				logger.Println("Error loading plugin: " + file.Name())
				logger.Println(err)
				continue
			} else {
				logger.Println("Loading plugin: " + file.Name())
			}
			u, err := plugin.Lookup("Utterances")
			if err != nil {
				logger.Println("Error loading Utterances []string from plugin file " + file.Name())
				logger.Println(err)
				continue
			} else {
				if _, ok := u.(*[]string); ok {
					logger.Println("Utterances []string in plugin " + file.Name() + " are OK")
				} else {
					logger.Println("Error: Utterances in plugin " + file.Name() + " are not of type []string")
					continue
				}
			}
			a, err := plugin.Lookup("Action")
			if err != nil {
				logger.Println("Error loading Action func from plugin file " + file.Name())
				continue
			} else {
				if _, ok := a.(func(string, string) (string, string)); ok {
					logger.Println("Action func in plugin " + file.Name() + " is OK")
				} else {
					logger.Println("Error: Action func in plugin " + file.Name() + " is not of type func(string, string) string")
					continue
				}
			}
			n, err := plugin.Lookup("Name")
			if err != nil {
				logger.Println("Error loading Name string from plugin file " + file.Name())
				continue
			} else {
				if _, ok := n.(*string); ok {
					logger.Println("Name string in plugin " + *n.(*string) + " is OK")
				} else {
					logger.Println("Error: Name string in plugin " + file.Name() + " is not of type string")
					continue
				}
			}
			PluginUtterances = append(PluginUtterances, u.(*[]string))
			PluginFunctions = append(PluginFunctions, a.(func(string, string) (string, string)))
			PluginNames = append(PluginNames, *n.(*string))
			PluginList = append(PluginList, plugin)
			logger.Println(file.Name() + " loaded successfully")
		}
		// else {
		//	logger.Println("Not loading " + file.Name() + ". Plugins must be built with 'go build -buildmode=plugin' and must end in '.so'.")
		//}
	}
}
