package main

import (
	"github.com/Kingdo777/serverless.instance.select/pkg/instance"
	"github.com/Kingdo777/serverless.instance.select/pkg/k8s"
	apiv1 "k8s.io/api/core/v1"
)

func main() {
	clientSet := k8s.GetClientSet()
	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)
	serviceClient := clientSet.CoreV1().Services(apiv1.NamespaceDefault)
	nodesClient := clientSet.CoreV1().Nodes()

	imageName := "kingdo/autoscale-go"

	//创建资源
	deployment := k8s.CreateDeployment(deploymentsClient, imageName)
	svc := k8s.CreateService(serviceClient, deployment)

	//获取核心数据结构SI，这一步主要是运行各个实例获取在不同并发下延时
	SI := instance.RunToGetData(30, deploymentsClient, k8s.GetUrl(nodesClient, svc))
	//通过上一步的数据完善信息
	SI.CompleteSI()
	//生成预测数据文件
	SI.MakePredicate()
	//打印SI信息
	SI.PrintSI()

	//删除资源
	k8s.DeleteDeployment(deploymentsClient)
	k8s.DeleteService(serviceClient, deployment)
}
