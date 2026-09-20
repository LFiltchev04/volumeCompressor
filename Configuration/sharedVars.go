package Configuration

import (
	"fmt"
	"os"

	"github.com/ceph/go-ceph/cephfs"
	"github.com/go-yaml/yaml"
)


type ReadableConfig struct{
	MonIPs []string `yaml:"MonIPs"`
	ClusterName string `yaml:"ClusterName"`
	ClusterKey string `yaml:"ClusterKey"`
	FilesystemID string `yaml:"FilesystemID"`
	DistributionPV string `yaml:"DistributionPV"`
}

type GlobalConfig struct {
	BaseConf ReadableConfig
	GlobalMnt cephfs.MountInfo
}


func (c *GlobalConfig) InitConfig(pathToConf string) error {
	byteSlop, err := os.ReadFile(pathToConf)
	
	var externConf ReadableConfig
	err = yaml.Unmarshal(byteSlop, &externConf)
	if err != nil {
		println("///====> Errored out when attempting to open config directory ", err.Error() )
	}
	c.BaseConf = externConf
	fmt.Printf("externConf: %v\n", externConf)
	glbMnt, err := cephfs.CreateMount()
	if err != nil {
		println("///====> Errored out when attempting to open config directory ", err.Error())
		return err
	}

	for _, monip  := range c.BaseConf.MonIPs {
		println("monIp from file: ", monip)
		glbMnt.SetConfigOption("mon_host", monip)
	}

	glbMnt.SetConfigOption("name", externConf.ClusterName)
	glbMnt.SetConfigOption("key", externConf.ClusterKey)
	glbMnt.Mount()
	
	c.GlobalMnt = *glbMnt
	return nil
}

var Global GlobalConfig

