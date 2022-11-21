package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"
	"fmt"
	// "context"

	// "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	klient "vm_crd/pkg/client/clientset/versioned"
	kInfFac "vm_crd/pkg/client/informers/externalversions"
	"vm_crd/pkg/controller"
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Printf("Building config from flags failed, %s, trying to build inclusterconfig", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Println("h1")
			log.Printf("error %s building inclusterconfig", err.Error())
		}
	}

	klientset, err := klient.NewForConfig(config)
	if err != nil {
		fmt.Println("h2")
		log.Printf("getting klient set %s\n", err.Error())
	}

	// client, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	log.Printf("getting std client %s\n", err.Error())
	// }

	fmt.Println(klientset)
	// klusters, err := klientset.SamplecontrollerV1alpha1().VMs("").List(context.Background(),metav1.ListOptions{})
	// if err != nil {
	// 	fmt.Println("h2")
	// 	log.Printf("listing errors %s\n", err.Error())
	// }

	// fmt.Printf("len of vms is: %d and name is %s\n",len(klusters.Items),klusters.Items[0].Name)
	
	
	infoFactory := kInfFac.NewSharedInformerFactory(klientset, 20*time.Minute)
	ch := make(chan struct{})
	c := controller.NewController(klientset, infoFactory.Samplecontroller().V1alpha1().VMs())
	
	go c.UpdateCpuUtilization("default")
	
	infoFactory.Start(ch)
	if err := c.Run(ch); err != nil {
		log.Printf("error running controller %s\n", err.Error())
	}
}