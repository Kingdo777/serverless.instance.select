package main

import (
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	"github.com/Kingdo777/serverless.instance.select/pkg/svm"
	"math"
	"strconv"
)

func main() {

	//for i := 0; i < len(config.VmConfigList); i++ {
	//	svm.Train(config.TrainDataFilePath + ".vm" + strconv.Itoa(i))
	//}

	var conc float64
	x := make(map[int]float64)
	for i := 0; i < len(config.VmConfigList); i++ {
		for j := 0; j < len(config.TargetLatency); j++ {
			x[0] = config.TargetLatency[j]
			conc = svm.Predicting("data/train.vm"+strconv.Itoa(i)+".model", x)
			svm.MakeTrainData(int(math.Ceil(conc)), config.TargetLatency[j], config.TrainDataFilePath+".vm"+strconv.Itoa(i)+".predicate")
		}
	}
}
