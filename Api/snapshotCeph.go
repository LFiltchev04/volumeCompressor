package Api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"compressor/Configuration"
)
func SnapshotCeph(snapshotName string, volumeName string)error{

	config := Configuration.Global.BaseConf.MonIPs[0]

	type snapshotBody struct {
		Vol_name string `json:"vol_name"`
		Subvol_name string `json:"subvol_name"`
		Snap_name string `json:"snap_name"`
	}

	var body snapshotBody
	body.Vol_name = "csi"
	body.Subvol_name = volumeName
	body.Snap_name = snapshotName
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "http://" + config + "/snapshotState", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}


	return nil
}