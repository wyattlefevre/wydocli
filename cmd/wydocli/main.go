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
		err := data.CleanTasks()
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	data.PrintTasks()
	err = data.CleanTasks()
	if err != nil {
		fmt.Println(err)
	}
}
