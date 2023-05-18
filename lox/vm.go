package lox

import "fmt"

const STACK_MAX = 256

type Vm struct {
	chunk    *Chunk
	ip       int
	stack    [STACK_MAX]Value
	stackTop int
}

type InterpretResult int

const (
	InterpretOk InterpretResult = iota
	InterpretCompileError
	InterpretRuntimeError
)

func NewVm(chunk *Chunk) *Vm {
	return &Vm{
		chunk:    chunk,
		ip:       0,
		stack:    *new([STACK_MAX]Value),
		stackTop: 0,
	}
}

func (vm *Vm) Interpret(source string) InterpretResult {
	compile(source)
	return InterpretOk
}

func (vm *Vm) push(value Value) {
	vm.stack[vm.stackTop] = value
	vm.stackTop += 1
}

func (vm *Vm) pop() Value {
	vm.stackTop -= 1
	return vm.stack[vm.stackTop]
}

func (vm *Vm) run() InterpretResult {
	for {
		// Only enable in debug
		vm.debugTraceExecution()

		instruction := vm.readByte()
		switch OpCode(instruction) {
		case OpReturn:
			{
				printValue(vm.pop())
				fmt.Println()
				return InterpretOk
			}
		case OpConstant:
			{
				constant := vm.readConstant()
				vm.push(constant)
			}
		case OpNegate:
			vm.push(-vm.pop())
		case OpAdd:
			vm.binaryOp('+')
		case OpSubtract:
			vm.binaryOp('-')
		case OpMultiply:
			vm.binaryOp('*')
		case OpDivide:
			vm.binaryOp('/')
		}
	}
}

func (vm *Vm) binaryOp(op byte) {
	b := vm.pop()
	a := vm.pop()
	switch op {
	case '+':
		vm.push(a + b)
	case '-':
		vm.push(a - b)
	case '*':
		vm.push(a * b)
	case '/':
		vm.push(a / b)
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
	for value := 0; value < vm.stackTop; value++ {
		fmt.Print("[ ")
		printValue(vm.stack[value])
		fmt.Print(" ]")
	}
	fmt.Println()
	vm.chunk.disassembleInstruction(vm.ip)
}
