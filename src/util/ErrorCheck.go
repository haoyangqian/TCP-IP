package util

import "fmt"

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
