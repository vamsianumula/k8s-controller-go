package controller

import (
	"context"
	"log"
	"time"
	do "vm_crd/pkg/do"

	"k8s.io/apimachinery/pkg/labels"

	"vm_crd/pkg/apis/samplecontroller.k8s.io/v1alpha1"
	klientset "vm_crd/pkg/client/clientset/versioned"

	kinf "vm_crd/pkg/client/informers/externalversions/samplecontroller.k8s.io/v1alpha1"
	klister "vm_crd/pkg/client/listers/samplecontroller.k8s.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client kubernetes.Interface

	// clientset for custom resource kluster
	klient klientset.Interface
	// kluster has synced
	klusterSynced cache.InformerSynced
	// lister
	kLister klister.VMLister
	// queue
	wq workqueue.RateLimitingInterface

}

func NewController(klient klientset.Interface, klusterInformer kinf.VMInformer) *Controller {


	c := &Controller{
		klient:        klient,
		klusterSynced: klusterInformer.Informer().HasSynced,
		kLister:       klusterInformer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
	}

	klusterInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, c.klusterSynced); !ok {
		log.Println("cache was not sycned")
	}

	go wait.Until(c.worker, time.Second, ch)

	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		// logs as well
		return false
	}

	defer c.wq.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error %s calling Namespace key func on cache for item", err.Error())
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name, error %s\n", err.Error())
		return false
	}

	kluster, err := c.kLister.VMs(ns).Get(name)

	if err != nil {
		log.Printf("error %s, Getting the kluster resource from lister", err.Error())
		return false
	}

	if kluster.DeletionTimestamp !=nil{
		statusCode, err:=do.Delete(kluster.Status.VmId)
		log.Printf("status-code : %d", statusCode)
		if err!=nil{
			log.Printf("error %s, deleting the cluster", err.Error())
			return false
		}
		if statusCode==204 || statusCode==404{
			kluster.Finalizers= make([]string, 0)
			// log.Printf("Failed to delete VM through API with status code: %d", statusCode)
		}
		return true
	}

	log.Printf("kluster spec that we have is %+v\n", kluster.Status)

	res, err := do.Check(kluster.Name)
	if err != nil{
		log.Printf("Failed to check name with error %s", err.Error())
		return false
	}

	if res!=200{
		log.Printf("Forbidden name, deleting the vm")
		err2:=c.deleteVm(name, kluster)
		if err2!=nil{
			log.Printf("error %s, deleting the vm", err2.Error())
		}
		return false
	}

	id, err := do.Create(kluster.Name)
	if err != nil {
		log.Printf("errro %s, creating the cluster", err.Error())
	}
	log.Printf("Request to create VM successful")

	err = c.updateStatusVmId(id, kluster)
	if err != nil {
		log.Printf("error %s, updating status of the kluster %s\n", err.Error(), kluster.Name)
	}

	return true
}
func (c *Controller) handleAdd(obj interface{}) {
	log.Println("handleAdd was called")
	c.wq.Add(obj)
}

func (c *Controller) handleDel(obj interface{}) {
	log.Println("handleDel was called")
	c.wq.Add(obj)
}

func (c *Controller) deleteVm(name string, kluster  *v1alpha1.VM) error {
	err:= c.klient.SamplecontrollerV1alpha1().VMs(kluster.Namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err!=nil{
		runtime.HandleError(err)
		log.Printf("delete vm failed with err %s\n",err)
		return err
	}
	return nil
}

func (c *Controller) updateStatusVmId(id string, kluster *v1alpha1.VM) error {
	// get the latest version of kluster
	k, err := c.klient.SamplecontrollerV1alpha1().VMs(kluster.Namespace).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	log.Printf("kluster spec that we have is %+v\n", kluster.Spec)
	k.Status.VmId= id
	_, err = c.klient.SamplecontrollerV1alpha1().VMs(kluster.Namespace).UpdateStatus(context.Background(), k, metav1.UpdateOptions{})
	return err
}

func (c *Controller) UpdateCpuUtilization(ns string) error{
	log.Printf("Starting update thread")
	for{
		selector:= labels.NewSelector()
		vmList, err :=c.kLister.VMs(ns).List(selector)
		if err!=nil{
			log.Printf("")
			return err
		}
		for _, vm := range vmList{
			cpuUtilization, err2 := do.GetCpu(vm.Status.VmId)
			if err2!=nil{
				log.Printf("Failed to get cpu")
			}
			vm.Status.CpuUtilization=cpuUtilization
			
			_, err2 = c.klient.SamplecontrollerV1alpha1().VMs(ns).UpdateStatus(context.Background(), vm, metav1.UpdateOptions{})
			if err2!=nil{
				log.Printf("Failed to update cpu")
			}
		}
		log.Printf("Update done")
		time.Sleep(1*time.Minute)
		
	}
	return nil
}