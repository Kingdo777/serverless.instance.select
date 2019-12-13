package main

import (
//"fmt"
//"os"
)

//const (
//	TrainDataFilePath = "data/train"
//)

func maini() {
	for i := 0; i < 100; i++ {
		print(TrainDataFilePath + ".vm")
		makeTrainData(i, 0.1*float64(i), TrainDataFilePath+".vm")
	}
}

//func makeTrainData(conc int, latency float64, filename string) {
//	//每执行一次，添加一次
//	fp, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 6)
//	defer fp.Close()
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fp.WriteString(fmt.Sprintf("%d %f\n", conc, latency))
//}
