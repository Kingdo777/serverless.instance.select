package k8s

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	"strconv"
)

func CreateService(serviceClient v12.ServiceInterface, deployment *appsv1.Deployment) *apiv1.Service {

	//不管3721先删除一下
	//deletePolicy := metav1.DeletePropagationForeground
	//serviceClient.Delete(deployment.Name, &metav1.DeleteOptions{
	//	PropagationPolicy: &deletePolicy,
	//})

	// Create a Service named "my-service" that targets "pod-group":"my-pod-group"
	port := deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort
	svc, err := serviceClient.Create(&apiv1.Service{
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

func DeleteService(serviceClient v12.ServiceInterface, deployment *appsv1.Deployment) {
	fmt.Println("Deleting svc...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := serviceClient.Delete(deployment.Name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted svc.")
}

func GetUrl(nodesClient v12.NodeInterface, svc *apiv1.Service) string {
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
	//url := address + ":" + nodePort + "?prime=10"
	url := address + ":" + nodePort
	fmt.Println(url)
	return url
}
