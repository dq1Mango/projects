package main

import "fmt"

func funMath(num int) {
	fmt.Println((num - 10) & 6)
}

func man() {
	funMath(13)
	funMath(14)
}
