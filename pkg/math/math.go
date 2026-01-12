package math

import (
	"dicer/pkg/logger"
	"dicer/pkg/stack"
	"errors"
	"strconv"
	"strings"
)

/*****************************************
* Postfix expression evaluation
* Example: EvaluatePostfixExpression("64+")
*****************************************/
func IsOperator(input string) bool {
	var operators = map[string]bool{
		"+": true,
		"-": true,
		"*": true,
		"/": true,
	}

	_, found := operators[input]
	return found
}

func IsOperand(input string) bool {
	_, err := strconv.Atoi(input)
	if err == nil {
		return true
	}
	return false
}

func EvaluatePostfixExpression(input string) (int, error) {
	var stack = &stack.StackList[int]{}
	// var stack = structures.CreateArrayStack[int]()

	slice := strings.Split(input, " ")
	for _, item := range slice {
		num, err := strconv.Atoi(item)
		if err == nil {
			stack.Push(num)
		}

		found := IsOperator(item)
		if found {
			// Evaluate Operand 1
			operand2, err := stack.Top()
			if err != nil {
				return -1, errors.New("Unbalanced postfix expression")
			}
			stack.Pop()

			// Evaluate Operand 2
			operand1, err := stack.Top()
			if err != nil {
				return -1, errors.New("Unbalanced postfix expression")
			}
			stack.Pop()

			// Evaluate Operators
			switch item {
			case "+":
				value := operand1 + operand2
				stack.Push(value)
			case "-":
				value := operand1 - operand2
				stack.Push(value)
			case "*":
				value := operand1 * operand2
				stack.Push(value)
			case "/":
				value := operand1 / operand2
				stack.Push(value)
			}
		}
	}

	return stack.Top()
}

/********************************************************
* Converts an infix expression to postfix
* Example: InfixToPostfix("( 7 + 55 ) * ( 8 / 2 )")
********************************************************/
func InfixToPostfix(input string) string {
	var stack = &stack.StackList[string]{}
	var exp string

	slice := strings.Split(input, " ")
	for _, item := range slice {
		if IsOperand(item) {
			exp = exp + " " + item
		} else if IsOperator(item) {
			for !stack.IsEmpty() && !isOpeningParenthesisOnStack(stack) && hasEqualOrHigherPrecedence(stack, item) {
				stackTop, err := stack.Top()
				if err != nil {
					logger.LogError(err.Error())
				}

				exp = exp + " " + stackTop
				stack.Pop()
			}

			stack.Push(item)
		} else if item == "(" {
			stack.Push(item)
		} else if item == ")" {
			for !stack.IsEmpty() && !isOpeningParenthesisOnStack(stack) {
				stackTop, err := stack.Top()
				if err != nil {
					logger.LogError(err.Error())
				}
				exp = exp + " " + stackTop
				stack.Pop()
			}
			stack.Pop()
		}
	}

	for !stack.IsEmpty() {
		stackTop, err := stack.Top()
		if err != nil {
			logger.LogError(err.Error())
		}
		exp = exp + " " + stackTop
		stack.Pop()
	}

	return exp
}

func isOpeningParenthesisOnStack(stack *stack.StackList[string]) bool {
	onStack, err := stack.Top()
	if err != nil {
		return false
	}
	return onStack == "("
}

func hasEqualOrHigherPrecedence(stack *stack.StackList[string], current string) bool {
	var operators = map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	onStack, err := stack.Top()
	if err != nil {
		return false
	}

	onStackPrec := operators[onStack]
	currentPrec := operators[current]

	return onStackPrec >= currentPrec
}

func IsBalancedParens(input string) bool {
	var stack = stack.CreateArrayStack[rune]()

	for _, runeValue := range input {
		switch runeValue {
		case '{', '(', '[':
			stack.Push(runeValue)
		case '}', ')', ']':

			openRune, _ := stack.Top()

			if stack.IsEmpty() || (openRune == '{' && runeValue != '}' || openRune == '[' && runeValue != ']' || openRune == '(' && runeValue != ')') {
				return false
			} else {
				stack.Pop()
			}
		}
	}

	return stack.IsEmpty()
}
