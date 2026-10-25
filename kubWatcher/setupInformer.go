package kubwatcher

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type InformerConfig struct {
	clientSet *kubernetes.Clientset
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
	ic.clientSet = clientSet

	return nil
}

// verifies a namespace exists
func (ic *InformerConfig) NamespaceExists(ns string) (bool, error) {

	ctx := context.TODO()

	//verifies namespace
	_, err := ic.clientSet.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	//verifies headless service
	_, err = ic.clientSet.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}


	//verifies PVC exists, only one PV per environment
	_, err = ic.clientSet.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (ic *InformerConfig) CreateNamespace(ns string) error {
	ctx := context.TODO()

	//creates NS
	_, err := ic.clientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}

		log.Println("err creating ns: ", err)
		return err
	}

	log.Println("Created namespace", ns)




	//creates the headless service for the ssh bind

	//the naming convention for the service connection
	//the entire ssh flow depends on this thing
	svcName := "svc-" + ns


	svc := &corev1.Service{
    	ObjectMeta: metav1.ObjectMeta{
        	Name:      svcName,
        	Namespace: ns,
    	},
    	Spec: corev1.ServiceSpec{
        	ClusterIP: "None", 
        	Selector: map[string]string{
            	"type": "workspace",
        	},
    	},
	}

	
	_, err = ic.clientSet.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {

		if apierrors.IsAlreadyExists(err) {
			return nil
		}

		log.Println("err creating svc: ", err)
		return err
	}



	//allocates the PVC from the ceph, storageclass is assumed

	cephfsSc := "cephfs-sc"
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			StorageClassName: &cephfsSc,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}

	_, err = ic.clientSet.CoreV1().PersistentVolumeClaims(ns).Create(ctx, &pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
