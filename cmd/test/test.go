package main

import "fmt"

type M [10]map[int]float64

func main() {
	for i := 0.01; i < 0.03; i += 0.001 {
		fmt.Printf("%.3f,", i)
	}

}
