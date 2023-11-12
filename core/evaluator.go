package core

import "fmt"

func (i *IntegerNode) Eval(scope *InterpreterScope) int {
	return i.value
}

var BuiltinFuncMap = map[string]func([]int) int{
	"+": builtinAdd,
	"-": builtinSub,
	"*": builtinMul,
	"/": builtinDiv,
}

func builtinAdd(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum += number
	}
	return sum
}

func builtinSub(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum -= number
	}
	return sum
}

func builtinMul(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum *= number
	}
	return sum
}

func builtinDiv(nums []int) int {
	sum := nums[0]
	for _, number := range nums[1:] {
		sum /= number
	}
	return sum
}

func evalBuiltin(operand string, arguments []ASTNode, scope *InterpreterScope) int {
	evaluatedArgs := make([]int, 0)
	for _, arg := range arguments {
		evaluatedArg := arg.Eval(scope)
		evaluatedArgs = append(evaluatedArgs, evaluatedArg)
	}
	return BuiltinFuncMap[operand](evaluatedArgs)
}

func applyFunction(function *FunctionNode, functionEnv *InterpreterScope) int {
	value := 0
	for _, expr := range function.body {
		value = expr.Eval(functionEnv)
	}
	return value
}

func (s *SExpr) Eval(scope *InterpreterScope) int {
	if Includes(builtInOperations, s.operand) {
		return evalBuiltin(s.operand, s.arguments, scope)
	}
	if scope.get(s.operand) == nil {
		panic(fmt.Sprintf("%s not in scope", s.operand))
	}
	astNode := scope.get(s.operand)
	function, ok := astNode.(*FunctionNode)
	if !ok {
		panic("Expected function node,got some nonsense")
	}
	extendedEnv := NewInterpreterScope(scope)
	for indx := range s.arguments {
		extendedEnv.inner[function.arguments[indx]] = &IntegerNode{value: s.arguments[indx].Eval(scope)}
	}

	return applyFunction(function, extendedEnv)
}

func (f *FunctionNode) Eval(scope *InterpreterScope) int {
	scope.inner[f.name] = f
	value := 0
	if f.name == "main" {
		for _, expr := range f.body {
			value = expr.Eval(scope)
		}
	}
	return value
}

func (r *ReferenceNode) Eval(scope *InterpreterScope) int {
	panic("Interpreter does not support references")
}

func (s *InterpreterScope) get(variable string) ASTNode {
	if s.inner[variable] == nil {
		if s.outer == nil {
			return nil
		}
		return s.outer.get(variable)
	}
	return s.inner[variable]
}

func (s *CompilerScope) get(variable string) (string, error) {
	inner, innerOk := s.inner[variable]
	if !innerOk {
		if s.outer == nil {
			return "", fmt.Errorf("variable not defined")
		}
		return s.outer.get(variable)
	}
	return inner, nil
}

func (i *IdentifierNode) Eval(scope *InterpreterScope) int {
	if scope.get(i.name) == nil {
		panic(fmt.Sprintf("Compiler error: %s not found in scope", i.name))
	}
	return scope.get(i.name).Eval(scope)
}

var (
	comparisionOps = []string{"<", ">", "="}
	arithmeticOps  = []string{"+", "-", "*", "/", "%"}
	systemCalls    = []string{"sys_write"}
)

func (i *IfNode) Eval(scope *InterpreterScope) int {
	if !Includes(comparisionOps, i.condition.operand) {
		panic("Should have a comparision operator in if condition")
	}
	if len(i.condition.arguments) != 2 {
		panic("Conditional operators are binary")
	}
	var isCondTrue bool
	switch i.condition.operand {
	case "<":
		isCondTrue = i.condition.arguments[0].Eval(scope) < i.condition.arguments[1].Eval(scope)
		break
	case ">":
		isCondTrue = i.condition.arguments[0].Eval(scope) > i.condition.arguments[1].Eval(scope)
		break
	case "=":
		isCondTrue = i.condition.arguments[0].Eval(scope) == i.condition.arguments[1].Eval(scope)
		break
	default:
		panic("Error")
	}
	if isCondTrue {
		return i.trueExpr.Eval(scope)
	} else {
		if i.falseExpr != nil {
			return i.falseExpr.Eval(scope)
		}
		return 0
	}
}

func getMidpointIndex[T any](arr []T) int {
	if len(arr)%2 == 0 {
		return len(arr) / 2
	} else {
		return (len(arr) + 1) / 2
	}
}
