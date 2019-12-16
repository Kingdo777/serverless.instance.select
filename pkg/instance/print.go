package instance

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	"github.com/Kingdo777/serverless.instance.select/pkg/svm"
	"math"
	"strconv"
)

func (SI *ServiceInstance) PrintSI() {
	for vmIndex := 0; vmIndex < len(config.VmConfigList); vmIndex++ {
		fmt.Printf("vm%d:maxConc--->%d\n", vmIndex, SI.InstanceRunModel[vmIndex].MaxConcurrency)
	}
	for concIndex := 0; concIndex < len(config.Concurrency); concIndex++ {
		fmt.Printf("conc.%d:bestVM.cpu--->%d\n", config.Concurrency[concIndex], SI.ConcurrencyInstance[concIndex].Cpu)
	}
}

func (SI *ServiceInstance) MakePredicate() {
	var conc float64
	x := make(map[int]float64)
	for vmIndex, vmInstance := range SI.InstanceRunModel {
		if vmInstance.IsWorked {
			for j := 0; j < len(config.TargetLatency); j++ {
				x[0] = config.TargetLatency[j]
				conc = svm.Predicting(vmInstance.Model, x)
				svm.MakeTrainData(int(math.Ceil(conc)), config.TargetLatency[j], config.TrainDataFilePath+".vm"+strconv.Itoa(vmIndex)+".predicate")
			}
		}
	}
}
