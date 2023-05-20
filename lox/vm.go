package lox

import (
	"fmt"
	"os"
)

type Vm struct {
	chunk *Chunk
	ip    int
	stack []Value
}

type InterpretResult int

const (
	InterpretOk InterpretResult = iota
	InterpretCompileError
	InterpretRuntimeError
)

func NewVm() *Vm {
	return &Vm{
		chunk: NewChunk(),
		ip:    0,
		stack: make([]Value, 0),
	}
}

func (vm *Vm) resetVm() {
	vm.stack = make([]Value, 0)
	vm.chunk = NewChunk()
	vm.ip = 0
}

func (vm *Vm) Interpret(source string) InterpretResult {
	compiler := NewCompiler(source, vm.chunk)
	if !compiler.compile() {
		vm.resetVm()
		return InterpretCompileError
	}
	return vm.run()
}

func (vm *Vm) push(value Value) {
	vm.stack = append(vm.stack, value)
}

func (vm *Vm) pop() Value {
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return value
}

func (vm *Vm) peek(distance int) Value {
	return vm.stack[len(vm.stack)-1-distance]
}

func (vm *Vm) run() InterpretResult {
	for {
		// Only enable in debug
		vm.debugTraceExecution()

		instruction := vm.readByte()
		switch OpCode(instruction) {
		case OpReturn:
			{
				vm.pop().print()
				fmt.Println()
				return InterpretOk
			}
		case OpConstant:
			{
				constant := vm.readConstant()
				vm.push(constant)
			}
		case OpNegate:
			if _, isNumber := vm.peek(0).(NumberValue); !isNumber {
				vm.runtimeError("Operand must be a number.")
				return InterpretRuntimeError
			}
			vm.push(-vm.pop().(NumberValue))
		case OpAdd, OpSubtract, OpMultiply, OpDivide, OpGreater, OpLess, OpGreaterEqual, OpLessEqual:
			{
				_, isBNumber := vm.peek(0).(NumberValue)
				_, isANumber := vm.peek(1).(NumberValue)
				if !isANumber || !isBNumber {
					vm.runtimeError("Operands must be numbers.")
					return InterpretRuntimeError
				}
				b := vm.pop().(NumberValue)
				a := vm.pop().(NumberValue)
				switch OpCode(instruction) {
				case OpAdd:
					vm.push(a + b)
				case OpSubtract:
					vm.push(a - b)
				case OpMultiply:
					vm.push(a * b)
				case OpDivide:
					vm.push(a / b)
				case OpGreater:
					vm.push(BoolValue(a > b))
				case OpLess:
					vm.push(BoolValue(a < b))
				case OpGreaterEqual:
					vm.push(BoolValue(a >= b))
				case OpLessEqual:
					vm.push(BoolValue(a <= b))
				}
			}
		case OpNil:
			vm.push(NilValue{})
		case OpTrue:
			vm.push(BoolValue(true))
		case OpFalse:
			vm.push(BoolValue(false))
		case OpNot:
			vm.push(BoolValue(!vm.pop().isTruthy()))
		case OpEqual:
			{
				b := vm.pop()
				a := vm.pop()
				vm.push(BoolValue(a == b))
			}
		case OpNotEqual:
			{
				b := vm.pop()
				a := vm.pop()
				vm.push(BoolValue(a != b))
			}
		}
	}
}

func (vm *Vm) readByte() byte {
	byte := vm.chunk.code[vm.ip]
	vm.ip += 1
	return byte
}

func (vm *Vm) readConstant() Value {
	return vm.chunk.constants[vm.readByte()]
}

func (vm *Vm) debugTraceExecution() {
	fmt.Print("          ")
	for value := range vm.stack {
		fmt.Print("[ ")
		vm.stack[value].print()
		fmt.Print(" ]")
	}
	fmt.Println()
	vm.chunk.disassembleInstruction(vm.ip)
}

func (vm *Vm) runtimeError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
	fmt.Printf("[line %d] in script\n", vm.chunk.lines[vm.ip-1])
	vm.resetVm()
}
