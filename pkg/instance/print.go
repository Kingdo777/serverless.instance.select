package instance

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
)

func PrintSI(SI ServiceInstance) {
	for vmIndex := 0; vmIndex < len(config.VmConfigList); vmIndex++ {
		fmt.Printf("vm%d:maxConc--->%d\n", vmIndex, SI.InstanceRunModel[vmIndex].MaxConcurrency)
	}
	for concIndex := 0; concIndex < len(config.Concurrency); concIndex++ {
		fmt.Printf("conc.%d:bestVM.cpu--->%d\n", config.Concurrency[concIndex], SI.ConcurrencyInstance[concIndex].Cpu)
	}
}
