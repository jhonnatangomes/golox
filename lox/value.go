package lox

import "fmt"

type Value float64

func printValue(value Value) {
	fmt.Printf("%g", value)
}
