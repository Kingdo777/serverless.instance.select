package main

import (
	"bufio"
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"path/filepath"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
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
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	nodesClient := clientset.CoreV1().Nodes()

	deployment := createDeployment(deploymentsClient)
	svc := createService(serviceClient, deployment)

	//latency := sendRequest(getUrl(nodesClient, svc), 10)
	//fmt.Printf("%f", latency)

	SI := runToGetData(30, deploymentsClient, getUrl(nodesClient, svc))
	completeSI(&SI)
	for vmIndex := 0; vmIndex < len(vmConfigList); vmIndex++ {
		fmt.Printf("vm%d:maxConc--->%d\n", vmIndex, SI.instanceRunModel[vmIndex].maxConcurrency)
	}
	for concIndex := 0; concIndex < len(concurrency); concIndex++ {
		fmt.Printf("conc.%d:bestVM.cpu--->%d\n", concurrency[concIndex], SI.concurrencyInstance[concIndex].cpu)
	}
	//val, _ := json.Marshal(SI)
	//fmt.Println(string(val))

	//for index, vm := range vmList() {
	//	updateDeployment(deploymentsClient, vm)
	//	latency := sendRequest(getUrl(nodesClient, svc), 10, 10)
	//	fmt.Printf("vm%d(cpu:%dm,mem:%dMi):%f\n", index, vmConfigList[index].cpu, vmConfigList[index].mem, latency)
	//	updateDeployment(deploymentsClient, vmInstanceDefault)
	//}

	//listDeployment(deploymentsClient)

	deleteDeployment(deploymentsClient)

}

func getUrl(nodesClient v12.NodeInterface, svc *apiv1.Service) string {
	//nodes, _ := nodesClient.Get("minikube", metav1.GetOptions{})
	nodeList, _ := nodesClient.List(metav1.ListOptions{})
	node := nodeList.Items[0]
	var address string
	for _, nodeAddress := range node.Status.Addresses {
		//EKS需要打开端口转发
		if nodeAddress.Type == apiv1.NodeExternalIP {
			address = "http://" + nodeAddress.Address
		}
	}
	if address == "http://" {
		//没有外部IP使用内部
		for _, nodeAddress := range node.Status.Addresses {
			if nodeAddress.Type == apiv1.NodeInternalIP {
				address = "http://" + nodeAddress.Address
			}
		}
	}

	if node.Name == "minikube" {
		address = "http://192.168.99.100" //minikube的问题，nodeport没办法直接访问
	}
	nodePort := strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
	url := address + ":" + nodePort
	fmt.Println(url)
	return url
}

func sendRequest(url string, concurrency int, dur int) float64 {
	latency := hey(url, concurrency, strconv.Itoa(dur)+"s")
	return latency
}

func createService(serviceClient v12.ServiceInterface, deployment *appsv1.Deployment) *apiv1.Service {
	svc, err := serviceClient.Get("instance-select", metav1.GetOptions{})
	if err == nil {
		return svc
	}
	// Create a Service named "my-service" that targets "pod-group":"my-pod-group"
	port := int32(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	svc, err = serviceClient.Create(&apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "instance-select",
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeNodePort,
			//Type:     apiv1.ServiceTypeLoadBalancer,
			Selector: deployment.Spec.Selector.MatchLabels,
			Ports: []apiv1.ServicePort{
				{
					Port: port,
					//TargetPort:deployment.Spec.Template.Spec.Containers[0].Ports[0].HostPort,
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return svc

}

func createDeployment(deploymentsClient v1.DeploymentInterface) *appsv1.Deployment {

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "instance-select",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(0),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "instance-select",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "instance-select",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "autoscale",
							Image: "kingdo/autoscale-go",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							//ImagePullPolicy: apiv1.PullIfNotPresent,
							Resources: vmInstanceDefault.res,
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	for true {
		if result.Status.AvailableReplicas != *result.Spec.Replicas {
			result, _ = deploymentsClient.Get(result.Name, metav1.GetOptions{})
			fmt.Printf("Wait All Pod Ready.\n")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return result
}

func updateDeployment(deploymentsClient v1.DeploymentInterface, vm VmInstance) {
	// Update Deployment
	//prompt()
	//fmt.Println("Updating deployment...")
	//    You have two options to Update() this Deployment:
	//
	//    1. Modify the "deployment" variable and call: Update(deployment).
	//       This works like the "kubectl replace" command and it overwrites/loses changes
	//       made by other clients between you Create() and Update() the object.
	//    2. Modify the "result" returned by Get() and retry Update(result) until
	//       you no longer get a conflict error. This way, you can preserve changes made
	//       by other clients between Create() and Update(). This is implemented below
	//			 using the retry utility package included with client-go. (RECOMMENDED)
	//
	// More Info:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get("instance-select", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = int32Ptr(vm.replicas)
		result.Spec.Template.Spec.Containers[0].Resources = vm.res
		//result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := deploymentsClient.Update(result)
		if updateErr == nil {
			for true {
				if result.Status.AvailableReplicas != *result.Spec.Replicas {
					result, _ = deploymentsClient.Get(result.Name, metav1.GetOptions{})
					fmt.Printf("Wait All Pod Ready.\n")
					time.Sleep(1 * time.Second)
				} else {
					break
				}
			}
		}
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
	//等待更新结束
	//time.Sleep(30 * time.Second)
	//fmt.Println("Updated deployment...")
}

func listDeployment(deploymentsClient v1.DeploymentInterface) {
	// List Deployments
	prompt()
	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}

func deleteDeployment(deploymentsClient v1.DeploymentInterface) {
	// Delete Deployment
	prompt()
	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete("instance-select", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func int32Ptr(i int32) *int32 { return &i }
