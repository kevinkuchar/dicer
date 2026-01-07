package stack

import "errors"

/************************************
* Stack Interface
************************************/
type Stack[T any] interface {
	Push(val T) error
	Pop()
	Top() (T, error)
	IsEmpty() bool
}

/*********************************
* Array Stack Implementation
*********************************/
const MAX_STACK_SIZE = 20

type ArrayStack[T any] struct {
	top    int
	values [MAX_STACK_SIZE]T
}

func CreateArrayStack[T any]() *ArrayStack[T] {
	var stack = &ArrayStack[T]{
		top:    -1,
		values: [MAX_STACK_SIZE]T{},
	}

	return stack
}

func (stack *ArrayStack[T]) Push(val T) error {
	if stack.top == MAX_STACK_SIZE-1 {
		return errors.New("stack is full")
	}
	stack.top++
	stack.values[stack.top] = val

	return nil
}

func (stack *ArrayStack[T]) Pop() {
	if stack.IsEmpty() {
		return
	}
	stack.top--
}

func (stack *ArrayStack[T]) Top() (T, error) {
	if stack.IsEmpty() {
		var zero T
		return zero, errors.New("stack is empty")
	}
	return stack.values[stack.top], nil
}

func (stack *ArrayStack[T]) IsEmpty() bool {
	return stack.top == -1
}

/*********************************
* Linked List Stack Implementation
**********************************/
type StackNode[T any] struct {
	Val  T
	Next *StackNode[T]
}

type StackList[T any] struct {
	Head *StackNode[T]
}

func (list *StackList[T]) Push(val T) error {
	new := &StackNode[T]{Val: val}
	new.Next = list.Head
	list.Head = new

	return nil
}

func (list *StackList[T]) Pop() {
	if list.Head == nil {
		return
	}
	currHead := list.Head
	newHead := currHead.Next

	list.Head = newHead
}

func (list *StackList[T]) Top() (T, error) {
	if list.Head == nil {
		var zero T
		return zero, errors.New("stack is empty")
	}

	return list.Head.Val, nil
}

func (list *StackList[T]) IsEmpty() bool {
	return list.Head == nil
}
