package main

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	"github.com/Kingdo777/serverless.instance.select/pkg/svm"
	"strconv"
)

func main() {
	for vmIndex := 0; vmIndex < len(config.VmConfigList); vmIndex++ {
		fmt.Println("Training " + ".vm" + strconv.Itoa(vmIndex) + " ...")
		modelFile := svm.Train("github.com/Kingdo777/serverless.instance.select/data/train.vm" + strconv.Itoa(vmIndex))
		fmt.Println("Trained ->>>> " + modelFile)
	}
}
