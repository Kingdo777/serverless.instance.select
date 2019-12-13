package main

import (
	"fmt"
	"github.com/ewalker544/libsvm-go"
	"os"
)

func makeTrainData(conc int, latency float64, filename string) {
	//每执行一次，添加一次
	fp, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 6)
	defer fp.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fp.WriteString(fmt.Sprintf("%d %f\n", conc, latency))
}

func svmTrain(trainDataFile string) string {
	param := libSvm.NewParameter() // Create a parameter object with default values
	param.KernelType = libSvm.POLY // Use the polynomial kernel

	model := libSvm.NewModel(param) // Create a model object from the parameter attributes

	// Create a problem specification from the training data and parameter attributes
	problem, _ := libSvm.NewProblem(trainDataFile, param)

	model.Train(problem) // Train the model from the problem specification

	modelFile := trainDataFile + ".model"

	model.Dump(modelFile)
	return modelFile
}

func svmPredicting(modelFile string) float64 {
	// Create a model object from the model file generated from training
	model := libSvm.NewModelFromFile(modelFile)

	x := make(map[int]float64)
	// Populate x with the test vector

	predictLabel := model.Predict(x) // Predicts a float64 label given the test vector

	return predictLabel
}
