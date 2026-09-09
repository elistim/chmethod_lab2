package numerics

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/scanner"
)

const DefaultFunction = "sqrt(x)+1"
const DefaultIntegrand = "1/(sqrt(1-x^2)*sqrt(1+x^2))"

// jet propagates a value and its first two derivatives by the chain rule.
type jet struct{ v, d, dd float64 }
type expression struct {
	op          string
	value       float64
	left, right *expression
	variable    bool
}
type expressionParser struct {
	scan  scanner.Scanner
	token rune
	count int
	err   error
}

func compileExpression(source string) (*expression, error) {
	if len(source) == 0 || len(source) > 256 {
		return nil, fmt.Errorf("формула должна содержать от 1 до 256 символов")
	}
	p := &expressionParser{}
	p.scan.Init(strings.NewReader(source))
	p.scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats
	p.scan.Error = func(_ *scanner.Scanner, message string) {
		p.err = fmt.Errorf("ошибка формулы: %s", message)
	}
	p.next()
	root := p.parse(0)
	if p.err != nil {
		return nil, p.err
	}
	if p.token != scanner.EOF {
		return nil, fmt.Errorf("лишний символ в формуле: %s", p.scan.TokenText())
	}
	root.markVariables()
	return root, nil
}
func (e *expression) markVariables() bool {
	e.variable = e.op == "x"
	if e.left != nil {
		e.variable = e.left.markVariables() || e.variable
	}
	if e.right != nil {
		e.variable = e.right.markVariables() || e.variable
	}
	return e.variable
}
func (p *expressionParser) next() { p.token = p.scan.Scan() }
func (p *expressionParser) parse(minimum int) *expression {
	p.count++
	if p.count > 100 {
		p.err = fmt.Errorf("формула слишком сложная")
		return &expression{}
	}
	var result *expression
	switch p.token {
	case '+', '-':
		op := string(p.token)
		p.next()
		result = &expression{op: "unary" + op, left: p.parse(25)}
	case '(':
		p.next()
		result = p.parse(0)
		if p.token != ')' {
			p.err = fmt.Errorf("в формуле не хватает закрывающей скобки")
		} else {
			p.next()
		}
	case scanner.Int, scanner.Float:
		v, err := strconv.ParseFloat(p.scan.TokenText(), 64)
		if err != nil || !finite(v) {
			p.err = fmt.Errorf("некорректное число в формуле")
		}
		result = &expression{value: v}
		p.next()
	case scanner.Ident:
		name := strings.ToLower(p.scan.TokenText())
		p.next()
		switch name {
		case "x":
			result = &expression{op: "x", variable: true}
		case "pi":
			result = &expression{value: math.Pi}
		case "e":
			result = &expression{value: math.E}
		case "sqrt", "sin", "cos", "tan", "exp", "ln", "log", "abs":
			if p.token != '(' {
				p.err = fmt.Errorf("после %s нужны скобки", name)
				return &expression{}
			}
			p.next()
			result = &expression{op: name, left: p.parse(0)}
			if p.token != ')' {
				p.err = fmt.Errorf("в функции %s не хватает закрывающей скобки", name)
			} else {
				p.next()
			}
		default:
			p.err = fmt.Errorf("неизвестное имя %q; используйте x, pi, e, sqrt, sin, cos, tan, exp, ln, log, abs", name)
			return &expression{}
		}
	default:
		p.err = fmt.Errorf("ожидается число, x или функция; найдено %q", p.scan.TokenText())
		return &expression{}
	}
	for p.err == nil {
		precedence := 0
		switch p.token {
		case '+', '-':
			precedence = 10
		case '*', '/':
			precedence = 20
		case '^':
			precedence = 30
		}
		if precedence == 0 || precedence < minimum {
			break
		}
		op := string(p.token)
		p.next()
		next := precedence + 1
		if op == "^" {
			next = precedence
		}
		result = &expression{op: op, left: result, right: p.parse(next)}
	}
	if result.left != nil {
		result.variable = result.variable || result.left.variable
	}
	if result.right != nil {
		result.variable = result.variable || result.right.variable
	}
	return result
}
func compose(a jet, v, d, dd float64) jet { return jet{v, d * a.d, dd*a.d*a.d + d*a.dd} }
func multiply(a, b jet) jet {
	return jet{a.v * b.v, a.d*b.v + a.v*b.d, a.dd*b.v + 2*a.d*b.d + a.v*b.dd}
}
func constantPower(a jet, p float64) jet {
	if p == 0 {
		return jet{v: 1}
	}
	if p == 1 {
		return a
	}
	d := p * math.Pow(a.v, p-1)
	dd := p * (p - 1) * math.Pow(a.v, p-2)
	return compose(a, math.Pow(a.v, p), d, dd)
}
func (e *expression) eval(x float64) jet {
	result := e.evaluate(x)
	if !e.variable {
		return jet{v: result.v}
	}
	return result
}
func (e *expression) evaluate(x float64) jet {
	if e.op == "" {
		return jet{v: e.value}
	}
	if e.op == "x" {
		return jet{v: x, d: 1}
	}
	a := e.left.eval(x)
	var b jet
	if e.right != nil {
		b = e.right.eval(x)
	}
	switch e.op {
	case "+":
		return jet{a.v + b.v, a.d + b.d, a.dd + b.dd}
	case "-":
		return jet{a.v - b.v, a.d - b.d, a.dd - b.dd}
	case "unary+":
		return a
	case "unary-":
		return jet{-a.v, -a.d, -a.dd}
	case "*":
		return multiply(a, b)
	case "/":
		return multiply(a, constantPower(b, -1))
	case "^":
		if !e.right.variable {
			return constantPower(a, b.v)
		}
		logarithm := compose(a, math.Log(a.v), 1/a.v, -1/(a.v*a.v))
		product := multiply(b, logarithm)
		v := math.Exp(product.v)
		return compose(product, v, v, v)
	case "sqrt":
		return constantPower(a, .5)
	case "sin":
		return compose(a, math.Sin(a.v), math.Cos(a.v), -math.Sin(a.v))
	case "cos":
		return compose(a, math.Cos(a.v), -math.Sin(a.v), -math.Cos(a.v))
	case "tan":
		v := math.Tan(a.v)
		return compose(a, v, 1+v*v, 2*v*(1+v*v))
	case "exp":
		v := math.Exp(a.v)
		return compose(a, v, v, v)
	case "ln", "log":
		return compose(a, math.Log(a.v), 1/a.v, -1/(a.v*a.v))
	case "abs":
		if a.v == 0 && e.variable {
			return jet{0, math.NaN(), math.NaN()}
		}
		if a.v < 0 {
			return jet{-a.v, -a.d, -a.dd}
		}
		return a
	}
	return jet{math.NaN(), math.NaN(), math.NaN()}
}
