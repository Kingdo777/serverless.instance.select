package svm

import (
	"fmt"
	"github.com/ewalker544/libsvm-go"
	"os"
)

func MakeTrainData(conc int, latency float64, filename string) {
	//每执行一次，添加一次
	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fp.Close()
	_, _ = fp.WriteString(fmt.Sprintf("%d 1:%f\n", conc, latency))
}

func Train(trainDataFile string) string {
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

func Predicting(modelFile string, x map[int]float64) float64 {
	// Create a model object from the model file generated from training
	model := libSvm.NewModelFromFile(modelFile)

	predictLabel := model.Predict(x) // Predicts a float64 label given the test vector

	return predictLabel
}
