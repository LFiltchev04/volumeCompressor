package Configuration

import (
	"fmt"
	"os"
	"log"

	"github.com/ceph/go-ceph/cephfs"
	"github.com/go-yaml/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type InformerConfig struct {
	ClientSet *kubernetes.Clientset
}

// has to be ran on init to start up the listener
func (ic *InformerConfig) SetupInformer(url string, kubeconfig string) error {
	log.Println("Setting up informer")

	config, err := clientcmd.BuildConfigFromFlags(url, kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
		return err
	}
	ic.ClientSet = clientSet

	return nil
}


type ReadableConfig struct{
	MonIPs []string `yaml:"MonIPs"`
	ClusterName string `yaml:"ClusterName"`
	ClusterKey string `yaml:"ClusterKey"`
	FilesystemID string `yaml:"FilesystemID"`
	DistributionPV string `yaml:"DistributionPV"`
	KubeconfigFile string `yaml:"KubeconfigFile"`
}

type GlobalConfig struct {
	BaseConf ReadableConfig
	GlobalMnt cephfs.MountInfo
	InformConfig InformerConfig
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

	err = c.InformConfig.SetupInformer("", c.BaseConf.KubeconfigFile)
	if err != nil {
		log.Println("err setting up informer", err)
	}
	
	c.GlobalMnt = *glbMnt
	return nil
}

var Global GlobalConfig
