package main

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	"github.com/Kingdo777/serverless.instance.select/pkg/instance"
	"github.com/Kingdo777/serverless.instance.select/pkg/k8s"
	apiv1 "k8s.io/api/core/v1"
)

func main() {
	clientset := k8s.GetClientSet()
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	nodesClient := clientset.CoreV1().Nodes()

	imageName := "kingdo/autoscale-go"

	//创建资源
	deployment := k8s.CreateDeployment(deploymentsClient, imageName)
	svc := k8s.CreateService(serviceClient, deployment)

	//获取核心数据结构SI，这一步主要是运行各个实例获取在不同并发下延时
	SI := instance.RunToGetData(30, deploymentsClient, k8s.GetUrl(nodesClient, svc))
	//通过上一步的数据完善信息
	instance.CompleteSI(&SI)

	//打印SI信息
	for vmIndex := 0; vmIndex < len(config.VmConfigList); vmIndex++ {
		fmt.Printf("vm%d:maxConc--->%d\n", vmIndex, SI.InstanceRunModel[vmIndex].MaxConcurrency)
	}
	for concIndex := 0; concIndex < len(config.Concurrency); concIndex++ {
		fmt.Printf("conc.%d:bestVM.cpu--->%d\n", config.Concurrency[concIndex], SI.ConcurrencyInstance[concIndex].Cpu)
	}

	//删除资源
	k8s.DeleteDeployment(deploymentsClient)
	k8s.DeleteService(serviceClient, deployment)
}
