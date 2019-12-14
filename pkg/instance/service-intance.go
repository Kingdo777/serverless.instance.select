package instance

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	"github.com/Kingdo777/serverless.instance.select/pkg/hey"
	"github.com/Kingdo777/serverless.instance.select/pkg/k8s"
	"github.com/Kingdo777/serverless.instance.select/pkg/svm"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
	"math"
	"os"
	"strconv"
	"time"
)

type RunModel struct {
	isWorked       bool
	MaxConcurrency int32
	model          string
}

type ServiceInstance struct {
	ConcurrencyInstance [len(config.Concurrency)]config.VmInstanceResourceCount
	InstanceRunModel    [len(config.VmConfigList)]RunModel
}

//SLO 是整数，单位是毫秒
func RunToGetData(SLO time.Duration, deploymentsClient v1.DeploymentInterface, url string) (SI ServiceInstance) {
	SecondSLO := float64(SLO) / 1000
	runTime := int(math.Ceil(SecondSLO * config.RuntimeMulity))
	for vmIndex, vm := range config.VmList() {
		_ = os.Remove(config.TrainDataFilePath + ".vm" + strconv.Itoa(vmIndex))
		k8s.UpdateDeployment(deploymentsClient, vm)
		var latency float64
		var conc int
		concurrencyIndex := 0
		for conc = config.Concurrency[concurrencyIndex]; conc <= config.Concurrency[len(config.Concurrency)-1]; {
			latency = hey.SendRequest(url, conc, runTime)
			fmt.Printf("request1:conc=%d and latency=%f\n", conc, latency)
			if latency < SecondSLO {
				svm.MakeTrainData(conc, latency, config.TrainDataFilePath+".vm"+strconv.Itoa(vmIndex))
				concurrencyIndex++
				if concurrencyIndex == len(config.Concurrency) {
					//此时最大的并发依然满足SLO
					break
				}
				conc = config.Concurrency[concurrencyIndex]
				SI.InstanceRunModel[vmIndex].isWorked = true
				//暂时等于latency
				//CostPerformanceTable[vmIndex][concurrencyIndex]=latency
			} else {
				if conc == 1 {
					//一个并发都无法接受，说明该实例无法用来运行改服务
					SI.InstanceRunModel[vmIndex].isWorked = false
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
			if concurrencyIndex < len(config.Concurrency) {
				//此时已经确定，最佳并发在concurrency[concurrencyIndex-1]到concurrency[concurrencyIndex]之间
				start := config.Concurrency[concurrencyIndex-1]
				end := config.Concurrency[concurrencyIndex] - 1
				fmt.Printf("bestConc between in %d and %d\n", start, end)
				for conc = (start + end) / 2; start < end; conc = (start + end) / 2 {
					if conc == start {
						latency = hey.SendRequest(url, end, runTime)
						fmt.Printf("request2:conc=%d and latency=%f\n", end, latency)
						if latency < SecondSLO {
							svm.MakeTrainData(end, latency, config.TrainDataFilePath+".vm"+strconv.Itoa(vmIndex))
							conc = end
						}
						break
					} else {
						latency = hey.SendRequest(url, conc, runTime)
						fmt.Printf("request3:conc=%d and latency=%f\n", conc, latency)
						if latency < SecondSLO {
							svm.MakeTrainData(conc, latency, config.TrainDataFilePath+".vm"+strconv.Itoa(vmIndex))
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
		SI.InstanceRunModel[vmIndex].MaxConcurrency = int32(conc)
		fmt.Printf("vm%d(cpu:%dm,mem:%dMi):bestConc=%d\n", vmIndex, config.VmConfigList[vmIndex].Cpu, config.VmConfigList[vmIndex].em, conc)
		if false {
			//TODO
			//这个地方要比较增加资源后，实例的响应时间时候在减小，如没有减小反而增加，那么已经没有继续测试的必要
		}
		k8s.UpdateDeployment(deploymentsClient, config.VmInstanceDefault)
	}
	return SI
}

func CompleteSI(SI *ServiceInstance) {
	makeCostPerformanceTable(SI)
	makeconcurrencyInstance(SI)
	makeModel(SI)
}

func makeModel(SI *ServiceInstance) {
	for vmIndex, vm := range SI.InstanceRunModel {
		fmt.Println("Training " + config.TrainDataFilePath + ".vm" + strconv.Itoa(vmIndex) + " ...")
		modelFile := svm.Train(config.TrainDataFilePath + ".vm" + strconv.Itoa(vmIndex))
		vm.model = modelFile
		fmt.Println("Trained ->>>> " + modelFile)
	}
}

func makeCostPerformanceTable(SI *ServiceInstance) {
	for vmIndex := 0; vmIndex < len(config.VmConfigList); vmIndex++ {
		if SI.InstanceRunModel[vmIndex].isWorked == false {
			for concIndex := 0; concIndex < len(config.Concurrency); concIndex++ {
				config.CostPerformanceTable[vmIndex][concIndex] = config.NotBest
				fmt.Printf("vmIndex.%d concIndex.%d value.%f\n", vmIndex, concIndex, config.CostPerformanceTable[vmIndex][concIndex])
			}
		} else {
			maxConc := SI.InstanceRunModel[vmIndex].MaxConcurrency
			for concIndex := 0; concIndex < len(config.Concurrency); concIndex++ {
				config.CostPerformanceTable[vmIndex][concIndex] = math.Ceil(float64(config.Concurrency[concIndex])/float64(maxConc)) * float64(config.VmCost()[vmIndex])
				fmt.Printf("vmIndex.%d concIndex.%d value.%f\n", vmIndex, concIndex, config.CostPerformanceTable[vmIndex][concIndex])
			}
		}
	}
}

func makeconcurrencyInstance(SI *ServiceInstance) {
	for concIndex := 0; concIndex < len(config.Concurrency); concIndex++ {
		SI.ConcurrencyInstance[concIndex] = minVMwithConc(concIndex)
	}
}

func minVMwithConc(concIndex int) config.VmInstanceResourceCount {
	index := 0
	for vmIndex := 1; vmIndex < len(config.VmConfigList); vmIndex++ {
		if config.CostPerformanceTable[vmIndex][concIndex] < config.CostPerformanceTable[index][concIndex] {
			index = vmIndex
		}
	}
	return config.VmConfigList[index]
}
