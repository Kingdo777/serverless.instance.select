package k8s

import (
	"bufio"
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
	"os"
	"time"
)

func CreateDeployment(deploymentsClient v1.DeploymentInterface, imageName string) *appsv1.Deployment {

	//不管3721先删除一下
	//deletePolicy := metav1.DeletePropagationForeground
	//deploymentsClient.Delete("instance-select", &metav1.DeleteOptions{
	//	PropagationPolicy: &deletePolicy,
	//})

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
							Image: imageName,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							//ImagePullPolicy: apiv1.PullIfNotPresent,
							Resources: config.VmInstanceDefault.Res,
						},
						{
							Name:  "measure",
							Image: "kingdo/instance-measure",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8081,
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

func UpdateDeployment(deploymentsClient v1.DeploymentInterface, vm config.VmInstance) {
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

		result.Spec.Replicas = int32Ptr(vm.Replicas)
		result.Spec.Template.Spec.Containers[0].Resources = vm.Res
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

func ListDeployment(deploymentsClient v1.DeploymentInterface) {
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

func DeleteDeployment(deploymentsClient v1.DeploymentInterface) {
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
