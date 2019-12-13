package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
	"math"
	"os"
	"strconv"
	"time"
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
		_ = os.Remove(TrainDataFilePath + ".vm" + strconv.Itoa(vmIndex))
		updateDeployment(deploymentsClient, vm)
		var latency float64
		var conc int
		concurrencyIndex := 0
		for conc = concurrency[concurrencyIndex]; conc <= concurrency[len(concurrency)-1]; {
			latency = sendRequest(url, conc, runTime)
			fmt.Printf("request1:conc=%d and latency=%f\n", conc, latency)
			if latency < SecondSLO {
				makeTrainData(conc, latency, TrainDataFilePath+".vm"+strconv.Itoa(vmIndex))
				concurrencyIndex++
				if concurrencyIndex == len(concurrency) {
					//此时最大的并发依然满足SLO
					break
				}
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
		if concurrencyIndex == 0 {
			//该实例无法处理任何一个请求
			conc = 0
		} else {
			if concurrencyIndex < len(concurrency) {
				//此时已经确定，最佳并发在concurrency[concurrencyIndex-1]到concurrency[concurrencyIndex]之间
				start := concurrency[concurrencyIndex-1]
				end := concurrency[concurrencyIndex] - 1
				fmt.Printf("bestConc between in %d and %d\n", start, end)
				for conc = (start + end) / 2; start < end; conc = (start + end) / 2 {
					if conc == start {
						latency = sendRequest(url, end, runTime)
						fmt.Printf("request2:conc=%d and latency=%f\n", end, latency)
						if latency < SecondSLO {
							makeTrainData(end, latency, TrainDataFilePath+".vm"+strconv.Itoa(vmIndex))
							conc = end
						}
						break
					} else {
						latency = sendRequest(url, conc, runTime)
						fmt.Printf("request3:conc=%d and latency=%f\n", conc, latency)
						if latency < SecondSLO {
							makeTrainData(conc, latency, TrainDataFilePath+".vm"+strconv.Itoa(vmIndex))
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
			} else {
				//此时最大的并发依然满足SLO
			}
		}
		SI.instanceRunModel[vmIndex].maxConcurrency = int32(conc)
		fmt.Printf("vm%d(cpu:%dm,mem:%dMi):bestConc=%d\n", vmIndex, vmConfigList[vmIndex].cpu, vmConfigList[vmIndex].mem, conc)
		if false {
			//TODO
			//这个地方要比较增加资源后，实例的响应时间时候在减小，如没有减小反而增加，那么已经没有继续测试的必要
		}
		updateDeployment(deploymentsClient, vmInstanceDefault)
	}
	return SI
}

func completeSI(SI *ServiceInstance) {
	makeCostPerformanceTable(SI)
	makeconcurrencyInstance(SI)

	time.Sleep(20 * time.Second)
	makeModel(SI)
}

func makeModel(SI *ServiceInstance) {
	for vmIndex, vm := range SI.instanceRunModel {
		fmt.Println("Training " + TrainDataFilePath + ".vm" + strconv.Itoa(vmIndex) + " ...")
		modelFile := svmTrain(TrainDataFilePath + ".vm" + strconv.Itoa(vmIndex))
		vm.model = modelFile
		fmt.Println("Trained ->>>> " + modelFile)
	}
}

func makeCostPerformanceTable(SI *ServiceInstance) {
	for vmIndex := 0; vmIndex < len(vmConfigList); vmIndex++ {
		if SI.instanceRunModel[vmIndex].isWorked == false {
			for concIndex := 0; concIndex < len(concurrency); concIndex++ {
				CostPerformanceTable[vmIndex][concIndex] = NotBest
				fmt.Printf("vmIndex.%d concIndex.%d value.%f\n", vmIndex, concIndex, CostPerformanceTable[vmIndex][concIndex])
			}
		} else {
			maxConc := SI.instanceRunModel[vmIndex].maxConcurrency
			for concIndex := 0; concIndex < len(concurrency); concIndex++ {
				CostPerformanceTable[vmIndex][concIndex] = math.Ceil(float64(concurrency[concIndex])/float64(maxConc)) * float64(vmCost()[vmIndex])
				fmt.Printf("vmIndex.%d concIndex.%d value.%f\n", vmIndex, concIndex, CostPerformanceTable[vmIndex][concIndex])
			}
		}
	}
}

func makeconcurrencyInstance(SI *ServiceInstance) {
	for concIndex := 0; concIndex < len(concurrency); concIndex++ {
		SI.concurrencyInstance[concIndex] = minVMwithConc(concIndex)
	}
}

func minVMwithConc(concIndex int) VmInstanceResourceCount {
	index := 0
	for vmIndex := 1; vmIndex < len(vmConfigList); vmIndex++ {
		if CostPerformanceTable[vmIndex][concIndex] < CostPerformanceTable[index][concIndex] {
			index = vmIndex
		}
	}
	return vmConfigList[index]
}
