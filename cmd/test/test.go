package main

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/tool"
)

func main() {
	fmt.Println(tool.Get("http://127.0.0.1:8081"))
}
