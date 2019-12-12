package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
	"math"
	"time"
)

const (
	RuntimeMulity = 50
	NotBest       = float64(99999999999)
)

var (
	concurrency = [...]int{
		1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024,
	}
	CostPerformanceTable = [len(vmConfigList)][len(concurrency)]float64{}
)

type InstanceRunModel struct {
	isWorked       bool
	maxConcurrency int32
	model          string
}

type ServiceInstance struct {
	concurrencyInstance [len(concurrency)]VmInstanceResourceCount
	instanceRunModel    [len(vmConfigList)]InstanceRunModel
}

//SLO 是整数，单位是毫秒
func runToGetData(SLO time.Duration, deploymentsClient v1.DeploymentInterface, url string) (SI ServiceInstance) {
	SecondSLO := float64(SLO) / 1000
	runTime := int(math.Ceil(SecondSLO * RuntimeMulity))
	for vmIndex, vm := range vmList() {
		updateDeployment(deploymentsClient, vm)
		var latency float64
		var conc int
		concurrencyIndex := 0
		for conc = concurrency[concurrencyIndex]; conc <= concurrency[len(concurrency)]; {
			latency = sendRequest(url, conc, runTime)
			if latency < SecondSLO {
				concurrencyIndex++
				conc = concurrency[concurrencyIndex]
				SI.instanceRunModel[vmIndex].isWorked = true
				//暂时等于latency
				//CostPerformanceTable[vmIndex][concurrencyIndex]=latency
			} else {
				if conc == 1 {
					//一个并发都无法接受，说明该实例无法用来运行改服务
					SI.instanceRunModel[vmIndex].isWorked = false
				}
				//for i := concurrencyIndex; i < len(concurrency); i++ {
				//	CostPerformanceTable[vmIndex][i] = NotBest
				//}
				break
			}
		}
		//此时已经确定，最佳并发在concurrency[concurrencyIndex-1]到concurrency[concurrencyIndex]之间
		start := concurrency[concurrencyIndex-1]
		end := concurrency[concurrencyIndex] - 1
		for conc = (start + end) / 2; start < end; conc = (start + end) / 2 {
			if conc == start {
				latency = sendRequest(url, end, runTime)
				if latency < SecondSLO {
					conc = end
				}
				break
			} else {
				latency = sendRequest(url, conc, runTime)
				if latency < SecondSLO {
					start = conc
				} else {
					if conc == end {
						conc = start
						break
					}
					end = conc - 1
				}
			}
		}
		SI.instanceRunModel[vmIndex].maxConcurrency = int32(conc)
		fmt.Printf("vm%d(cpu:%dm,mem:%dMi):bestConc=%d:latency=%f\n", vmIndex, vmConfigList[vmIndex].cpu, vmConfigList[vmIndex].mem, conc, latency)
		updateDeployment(deploymentsClient, vmInstanceDefault)
	}
	return SI
}
