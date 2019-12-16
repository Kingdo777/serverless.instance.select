package main

import (
	"github.com/Kingdo777/serverless.instance.select/pkg/config"
	"github.com/Kingdo777/serverless.instance.select/pkg/svm"
	"math"
	"strconv"
)

func main() {
	var conc float64
	x := make(map[int]float64)
	for j := 0; j < len(config.TargetLatency); j++ {
		x[0] = config.TargetLatency[j]
		conc = svm.Predicting("train.vm7.model", x)
		svm.MakeTrainData(int(math.Ceil(conc)), config.TargetLatency[j], config.TrainDataFilePath+".vm"+strconv.Itoa(7)+".predicate")
	}
}
