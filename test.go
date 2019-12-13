package main

func main1() {
	for i := 0; i < 100; i++ {
		makeTrainData(i, 0.1*float64(i), "test-train")
	}
}
