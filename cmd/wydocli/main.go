package main

import (
	"fmt"
	"github.com/wyattlefevre/wydocli/internal/data"
)

func main() {
	err := data.LoadAllTasks()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return
	}

	data.PrintProjects()
}
