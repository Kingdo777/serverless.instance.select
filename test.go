package main

//const (
//	TrainDataFilePath = "data/train"
//)

func main() {
	for i := 0; i < 100; i++ {
		makeTrainData(i, 0.1*float64(i), TrainDataFilePath+".vm")
	}
}

//func makeTrainData(conc int, latency float64, filename string) {
//	//每执行一次，添加一次
//	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer fp.Close()
//	fp.WriteString(fmt.Sprintf("%d %f\n", conc, latency))
//}
