package lox

import "fmt"

type Value interface {
	print()
	isTruthy() bool
}

type BoolValue bool

func (value BoolValue) print() {
	fmt.Print(value)
}

func (value BoolValue) isTruthy() bool {
	return bool(value)
}

type NilValue struct{}

func (NilValue) print() {
	fmt.Print("nil")
}

func (NilValue) isTruthy() bool {
	return false
}

type NumberValue float64

func (value NumberValue) print() {
	fmt.Printf("%g", value)
}

func (value NumberValue) isTruthy() bool {
	return true
}

type StringValue string

func (value StringValue) print() {
	fmt.Print(value)
}

func (value StringValue) isTruthy() bool {
	return true
}
