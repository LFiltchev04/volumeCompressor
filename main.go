package main

import (
	"compressor/Api"
	"compressor/Configuration"

) 
	


// i havent touched this repo in a while, i dont actually remember where i left it, it can carry out the needed logic but more than that is meh
func main() {

	var fixThis string
	//its the yaml config
	fixThis = "./srvConfig.yaml"
	
	Configuration.Global.InitConfig(fixThis)

	Api.Listener()	
}
