package main

import (
	"bufio"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
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

	deployment := createDeployment(deploymentsClient)

	svc := createService(serviceClient, deployment)

	nodesClient := clientset.CoreV1().Nodes()
	nodes, _ := nodesClient.Get("minikube", metav1.GetOptions{})
	nodeAddress := nodes.Status.Addresses[0].Address
	nodePort := strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
	url := nodeAddress + ":" + nodePort

	sendRequest(url)

	//updateDeployment(deploymentsClient)

	//listDeployment(deploymentsClient)

	deleteDeployment(deploymentsClient)

}

func sendRequest(url string) (latency float64) {

	fmt.Print(url)
	return 0
}

func createService(serviceClient v12.ServiceInterface, deployment *appsv1.Deployment) *apiv1.Service {
	svc, err := serviceClient.Get("instance-select", metav1.GetOptions{})
	if err == nil {
		return svc
	}
	// Create a Service named "my-service" that targets "pod-group":"my-pod-group"
	port := int32(8080)
	svc, err = serviceClient.Create(&apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "instance-select",
		},
		Spec: apiv1.ServiceSpec{
			Type:     apiv1.ServiceTypeNodePort,
			Selector: deployment.Labels,
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
			Replicas: int32Ptr(2),
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
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									apiv1.ResourceCPU:    *resource.NewMilliQuantity(125, resource.BinarySI),
									apiv1.ResourceMemory: *resource.NewMilliQuantity(128, resource.BinarySI),
								},
								Requests: apiv1.ResourceList{
									apiv1.ResourceCPU:    *resource.NewMilliQuantity(125, resource.BinarySI),
									apiv1.ResourceMemory: *resource.NewMilliQuantity(128, resource.BinarySI),
								},
							},
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
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	for true {
		if deployment.Status.AvailableReplicas != *deployment.Spec.Replicas {
			fmt.Printf("Wait All Pod Ready.\n")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return deployment
}

func updateDeployment(deploymentsClient v1.DeploymentInterface) {
	// Update Deployment
	prompt()
	fmt.Println("Updating deployment...")
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

		result.Spec.Replicas = int32Ptr(1) // reduce replica count
		//result.Spec.Template.Spec.Containers[0].Resources.Requests().Cpu().Set()
		//result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		deployment, updateErr := deploymentsClient.Update(result)
		if updateErr == nil {
			for true {
				if deployment.Status.AvailableReplicas != *deployment.Spec.Replicas {
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
	fmt.Println("Updated deployment...")
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
