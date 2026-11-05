package Configuration

import (
	"fmt"
	"log"
	"os"

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

type ReadableConfig struct {
	MonIPs         []string `yaml:"MonIPs"`
	ClusterName    string   `yaml:"ClusterName"`
	ClusterKey     string   `yaml:"ClusterKey"`
	FilesystemID   string   `yaml:"FilesystemID"`
	DistributionPV string   `yaml:"DistributionPV"`
	KubeconfigFile string   `yaml:"KubeconfigFile"`
}

type GlobalConfig struct {
	BaseConf     ReadableConfig
	GlobalMnt    cephfs.MountInfo
	InformConfig InformerConfig
}

func (c *GlobalConfig) InitConfig(pathToConf string) error {
	byteSlop, err := os.ReadFile(pathToConf)

	var externConf ReadableConfig
	err = yaml.Unmarshal(byteSlop, &externConf)
	if err != nil {
		println("///====> Errored out when attempting to open config directory ", err.Error())
	}
	c.BaseConf = externConf
	fmt.Printf("externConf: %v\n", externConf)
	// "id" is the entity without the "client." prefix
	glbMnt, err := cephfs.CreateMountWithId("kubernetes-csi")
	if err != nil {
		println("///====> Errored out when attempting to open config directory ", err.Error())
		return err
	}

	for _, monip := range c.BaseConf.MonIPs {
		println("monIp from file: ", monip)
		//glbMnt.SetConfigOption("mon_host", monip)

	}



	println("ceph auth, uname: ", externConf.ClusterName)
	err = glbMnt.ReadConfigFile("/etc/ceph/ceph.conf")
	if err != nil {
		println("err reading config", err.Error())
		panic(err)
	}

	err = glbMnt.SetConfigOption("keyring", "/etc/ceph/ceph.client.kubernetes-csi.keyring")
	if err != nil {
		println("err setting keyring", err.Error())
	}
	
	//glbMnt.SetConfigOption("key", externConf.ClusterKey)
	//glbMnt.SetConfigOption("auth_supported", "cephx")
	//glbMnt.SetConfigOption("mon_host", "[v2:192.168.1.40:3300]")
	//glbMnt.SetConfigOption("cephx_require_signatures", "false")
	//glbMnt.SetConfigOption("auth_supported", "cephx")


	//glbMnt.SetConfigOption("keyring", "/etc/ceph/ceph.client.kubernetes-csi.keyring")

	println("Setting cluster options")
	if err := glbMnt.Mount(); err != nil {
		println("mount error: ", err.Error())
	}
	if glbMnt.IsMounted() {
		println("Mounted successfully")
	} else {
		println("Mount failed")
	}

	err = c.InformConfig.SetupInformer("", c.BaseConf.KubeconfigFile)
	if err != nil {
		log.Println("err setting up informer", err)
	}

	c.GlobalMnt = *glbMnt
	return nil
}

var Global GlobalConfig
