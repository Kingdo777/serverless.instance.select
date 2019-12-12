package main

import "fmt"

func main() {
	var (
		conc   int
		target int
	)
	for true {
		start := 64
		end := 127
		fmt.Scanln(&target)
		for conc = (start + end) / 2; start < end; conc = (start + end) / 2 {
			if conc == start {
				break
			} else {
				if conc <= target {
					start = conc
				} else {
					if conc == end {
						conc = start
						break
					}
					end = conc - 1
				}
			}
		}
		fmt.Println(conc)
	}
}
