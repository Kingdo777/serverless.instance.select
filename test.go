package main

//const (
//	TrainDataFilePath = "data/train"
//)

func mai0n() {
	svmTrain("data/train.vm0")
	svmTrain("data/train.vm1")
}

//func svmTrain(trainDataFile string) string {
//	param := libSvm.NewParameter() // Create a parameter object with default values
//	param.KernelType = libSvm.POLY // Use the polynomial kernel
//
//	model := libSvm.NewModel(param) // Create a model object from the parameter attributes
//
//	// Create a problem specification from the training data and parameter attributes
//	problem, _ := libSvm.NewProblem(trainDataFile, param)
//
//	model.Train(problem) // Train the model from the problem specification
//
//	modelFile := trainDataFile + ".model"
//
//	model.Dump(modelFile)
//	return modelFile
//}
//
//func svmPredicting(modelFile string) float64 {
//	// Create a model object from the model file generated from training
//	model := libSvm.NewModelFromFile(modelFile)
//
//	x := make(map[int]float64)
//	// Populate x with the test vector
//
//	predictLabel := model.Predict(x) // Predicts a float64 label given the test vector
//
//	return predictLabel
//}
