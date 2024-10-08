package evaluator

import (
	"testing"

	"github.com/grantwforsythe/monkeylang/pkg/lexer"
	"github.com/grantwforsythe/monkeylang/pkg/object"
	"github.com/grantwforsythe/monkeylang/pkg/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"1 != 2", true},
		{`"HELLO" == "HELLO"`, true},
		{`"HELLO" == "WORLD"`, false},
		{`"HELLO" != "WORLD"`, true},
		{`"HELLO" != "HELLO"`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEvalBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)

	}
}

func TestEvalIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (0) { 10 }", nil},
		{"if (-1) { 10 }", nil},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (0) { 10 } else { 20 }", 20},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		value, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(value))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestEvalReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"fn(x, y){ x + y }(5, 5);", 10},
		{"fn(x, y){ return x + y }(5, 5);", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1) {
			  if (10 > 1) {
			    return true + false;
			  }

			  return 1;
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{"foobar", "identifier not found: foobar"},
		{`"Hello" - "World"`, "unknown operator: STRING - STRING"},
		{
			`{first([]): 1}`,
			"unhashable key: NULL",
		},
		{`{"a": 1}[fn(x) { x }]`, "unhashable key: FUNCTION"},
		{"1[0]", "index operator not supported: INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}

func TestEvalLetStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c", 15},
		{"let a = 5; let b = a - 5; b", 0},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		testIntegerObject(t, eval, tt.expected)
	}
}

func TestEvalFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	eval := testEval(input)

	fn, ok := eval.(*object.Function)
	if !ok {
		t.Fatalf("object is not of type *object.Function. got=%T (%+v)", eval, eval)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	if fn.Body.String() != "(x + 2)" {
		t.Fatalf("body is not %q. got=%q", "(x + 2)", fn.Body.String())
	}
}

func TestEvalFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
		{`let fib = fn(n) {
			if (n == 0) {
				return 0;
			}
			if (n == 1) {
				return 1;
			} else {
				return fib(n - 1) + fib(n - 2);
			}
		};
		fib(8);`, 21},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestEvalFunctionClosures(t *testing.T) {
	input := `
	let x = 100;

	let highFn = fn(x) {
		let y = 5;
		return fn(z) { x + y + z };
	};

	let closure = highFn(2);
	closure(10);
	`

	testIntegerObject(t, testEval(input), 17)
}

func TestEvalStringObject(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{

		{`"Hello, World!"`, "Hello, World!"},
		{`"Hello" + ", " + "World!"`, "Hello, World!"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestEvalBuiltinFunctionLen(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`len("")`, 0},
		{`len("Hello, World!")`, len("Hello, World!")},
		{`len(1)`, "argument to `len` not supported. got=INTEGER"},
		{`len()`, "wrong number of arguments. got=0, want=1"},
		{`len("1", "2")`, "wrong number of arguments. got=2, want=1"},
		{`len([])`, 0},
		{`len([1, 2, 3])`, 3},
		{`len([1, 2 * 5, 3])`, 3},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case string:
			err, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", eval, eval)
				continue
			}

			if err.Message != expected {
				t.Errorf("wrong error message. got=%s, expected=%s", err.Message, expected)
			}
		case *object.Null:
			testNullObject(t, eval)
		default:
			t.Errorf("object type is invalid. got=%T, expected=INTEGER, STRING, or NULL", eval)
		}
	}
}

func TestEvalBuiltinFunctionFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`first([1, 2, 3])`, 1},
		{`first([])`, NULL},
		{`first()`, "wrong number of arguments. got=0, want=1"},
		{`first(1)`, "'first' only accepts an array as an argument. got=INTEGER"},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case string:
			err, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", eval, eval)
				continue
			}

			if err.Message != expected {
				t.Errorf("wrong error message. got=%s, expected=%s", err.Message, expected)
			}
		case *object.Null:
			testNullObject(t, eval)
		default:
			t.Errorf("object type is invalid. got=%T, expected=INTEGER, STRING, or NULL", eval)
		}
	}
}

func TestEvalBuiltinFunctionLast(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`last([])`, NULL},
		{`last()`, "wrong number of arguments. got=0, want=1"},
		{`last(1)`, "'last' only accepts an array as an argument. got=INTEGER"},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, eval, int64(expected))
		case string:
			err, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", eval, eval)
				continue
			}

			if err.Message != expected {
				t.Errorf("wrong error message. got=%s, expected=%s", err.Message, expected)
			}
		case *object.Null:
			testNullObject(t, eval)
		default:
			t.Errorf("object type is invalid. got=%T, expected=INTEGER, STRING, or NULL", eval)
		}
	}
}

func TestEvalBuiltinFunctionRest(t *testing.T) {
	tests := []struct {
		input    string
		inital   any
		expected any
	}{
		{
			`rest([1, 2, 3])`,
			object.Array{
				Elements: []object.Object{
					&object.Integer{Value: 1},
					&object.Integer{Value: 2},
					&object.Integer{Value: 3},
				},
			},
			&object.Array{
				Elements: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}},
			},
		},
		{`rest([1])`,
			object.Array{Elements: []object.Object{&object.Integer{Value: 1}}},
			&object.Array{Elements: make([]object.Object, 0)}},
		{`rest([])`, object.Array{Elements: make([]object.Object, 0)}, NULL},
		// The initial value does not matter here as we are just checking the error message
		{`rest([1], [2])`, NULL, "wrong number of arguments. got=2, want=1"},
		{
			`rest(1)`,
			&object.Integer{Value: 1},
			"'rest' only accepts an array as an argument. got=INTEGER",
		},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case *object.Array:
			array, ok := eval.(*object.Array)
			if !ok {
				t.Errorf("object is not array. got=%T (%+v)", eval, eval)
				continue
			}

			if len(array.Elements) != len(expected.Elements) {
				t.Errorf("The number of elements does not match the expected %d. got=%d", len(expected.Elements), len(array.Elements))
				continue
			}

			// TODO: Refactor by implementing the compare interface so we can leverage the slice package
			for i := range len(array.Elements) {
				// This will always be of type int for testing purposes
				expectedValue := int64(expected.Elements[i].(*object.Integer).Value)
				testIntegerObject(t, array.Elements[i], expectedValue)
			}
		case string:
			err, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", eval, eval)
				continue
			}

			if err.Message != expected {
				t.Errorf("wrong error message. got=%s, expected=%s", err.Message, expected)
			}
		case *object.Null:
			testNullObject(t, eval)
		default:
			t.Errorf("object type is invalid. got=%T, expected=INTEGER, STRING, or NULL", eval)
		}
	}
}

func TestEvalBuiltinFunctionPush(t *testing.T) {
	tests := []struct {
		input    string
		inital   any
		expected any
	}{
		{
			`push([1], 2, 3)`,
			object.Array{
				Elements: []object.Object{&object.Integer{Value: 1}},
			},
			&object.Array{
				Elements: []object.Object{
					&object.Integer{Value: 1},
					&object.Integer{Value: 2},
					&object.Integer{Value: 3},
				},
			},
		},
		{
			`push([], 2, 3)`,
			object.Array{
				Elements: []object.Object{},
			},
			&object.Array{
				Elements: []object.Object{
					&object.Integer{Value: 2},
					&object.Integer{Value: 3},
				},
			},
		},
		{
			`push([1])`,
			object.Array{
				Elements: []object.Object{&object.Integer{Value: 1}},
			},
			"wrong number of arguments. got=1, want=>2",
		},
		{
			`push(1, 1)`,
			&object.Integer{Value: 1},
			"the first argument needs to be of type ARRAY. got=INTEGER",
		},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case *object.Array:
			array, ok := eval.(*object.Array)
			if !ok {
				t.Errorf("object is not array. got=%T (%+v)", eval, eval)
				continue
			}

			if len(array.Elements) != len(expected.Elements) {
				t.Errorf("The number of eval does not match the expected %d. got=%d", len(expected.Elements), len(array.Elements))
				continue
			}

			// TODO: Refactor by implementing the compare interface so we can leverage the slice package
			for i := range len(array.Elements) {
				// This will always be of type int for testing purposes
				expectedValue := int64(expected.Elements[i].(*object.Integer).Value)
				testIntegerObject(t, array.Elements[i], expectedValue)
			}
		case string:
			err, ok := eval.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", eval, eval)
				continue
			}

			if err.Message != expected {
				t.Errorf("wrong error message. got=%s, expected=%s", err.Message, expected)
			}
		case *object.Null:
			testNullObject(t, eval)
		default:
			t.Errorf("object type is invalid. got=%T, expected=INTEGER, STRING, or NULL", eval)
		}
	}
}

func TestEvalArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 4 - 5];"
	eval := testEval(input)

	result, ok := eval.(*object.Array)
	if !ok {
		t.Fatalf("result is not of type *object.Array. got=%T", eval)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("len(result) is not equal to 3. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], -1)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

// TODO: Test for duplicate key
func TestEvalHashLiterals(t *testing.T) {
	input := `let two = "two";
    {
        "one": 10 - 9,
        two: 1 + 1,
        "thr" + "ee": 6 / 2,
        4: 4,
        true: 5,
        false: 6
    }`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not of type *object.Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("result.Value is not equal to %d. got=%d", expected, result.Value)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not of type *object.Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("result.Value is not equal to %t. got=%t", expected, result.Value)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("obj is not NULL. got=%T (%+v)", obj, obj)
		return false
	}

	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not of type *object.String. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("result.Value is not equal to %s. got=%s", expected, result.Value)
		return false
	}

	return true
}
