package numerics

import (
	"fmt"
	"math"
	"strings"
)

type IntegrationInput struct {
	A          float64 `json:"a"`
	B          float64 `json:"b"`
	N          int     `json:"n"`
	Tolerance  float64 `json:"tolerance"`
	Method     string  `json:"method"`
	Order      int     `json:"order"`
	Expression string  `json:"expression"`
}
type Iteration struct {
	N        int      `json:"n"`
	H        float64  `json:"h"`
	Value    float64  `json:"value"`
	Estimate *float64 `json:"estimate"`
	Relative *float64 `json:"relative"`
}
type IntegrationResult struct {
	Input             IntegrationInput `json:"input"`
	Value             float64          `json:"value"`
	Reference         float64          `json:"reference"`
	AbsoluteError     float64          `json:"absoluteError"`
	RelativeError     float64          `json:"relativeError"`
	Estimate          float64          `json:"estimate"`
	N                 int              `json:"n"`
	Converged         bool             `json:"converged"`
	Message           string           `json:"message"`
	Iterations        []Iteration      `json:"iterations"`
	GaussNodes        []Point          `json:"gaussNodes"`
	Curve             []Point          `json:"curve"`
	HasReference      bool             `json:"hasReference"`
	AbsoluteCriterion bool             `json:"absoluteCriterion"`
}

func Integrand(x float64) float64 { return 1 / math.Sqrt((1-x*x)*(1+x*x)) }

// ReferencePrimitive uses the independently convergent binomial series
// (1-x^4)^(-1/2)=sum C(2k,k)/4^k*x^(4k). It is not a quadrature result.
func ReferencePrimitive(x float64) float64 {
	power := x
	coefficient := 1.0
	sum, correction := 0.0, 0.0
	x4 := x * x * x * x
	for k := 0; k < 100000; k++ {
		term := coefficient * power / float64(4*k+1)
		y := term - correction
		t := sum + y
		correction = (t - sum) - y
		sum = t
		if math.Abs(term)/(1-x4) < 2e-16*math.Max(1, math.Abs(sum)) {
			break
		}
		coefficient *= float64(2*k+1) / float64(2*k+2)
		power *= x4
	}
	return sum
}

// GaussRule returns Legendre roots and weights on [-1,1].
func GaussRule(order int) []Point {
	rule := make([]Point, order)
	for i := 0; i < (order+1)/2; i++ {
		z := math.Cos(math.Pi * (float64(i) + 0.75) / (float64(order) + 0.5))
		derivative := func(x float64) (float64, float64) {
			p0, p1 := 1.0, x
			for k := 2; k <= order; k++ {
				p0, p1 = p1, (float64(2*k-1)*x*p1-float64(k-1)*p0)/float64(k)
			}
			return p1, float64(order) * (x*p1 - p0) / (x*x - 1)
		}
		for k := 0; k < 100; k++ {
			p, d := derivative(z)
			next := z - p/d
			if math.Abs(next-z) < 2e-15 {
				z = next
				break
			}
			z = next
		}
		_, d := derivative(z)
		w := 2 / ((1 - z*z) * d * d)
		rule[i] = Point{-z, w}
		rule[order-1-i] = Point{z, w}
	}
	return rule
}

func quadrature(f func(float64) float64, a, b float64, n int, method string, rule []Point) float64 {
	h := (b - a) / float64(n)
	sum, correction := 0.0, 0.0
	add := func(v float64) { y := v - correction; t := sum + y; correction = (t - sum) - y; sum = t }
	switch method {
	case "midpoint":
		for i := 0; i < n; i++ {
			add(f(a + (float64(i)+0.5)*h))
		}
		return h * sum
	case "trapezoid":
		add(f(a) / 2)
		add(f(b) / 2)
		for i := 1; i < n; i++ {
			add(f(a + float64(i)*h))
		}
		return h * sum
	case "simpson":
		add(f(a))
		add(f(b))
		for i := 1; i < n; i++ {
			w := 2.0
			if i%2 == 1 {
				w = 4
			}
			add(w * f(a+float64(i)*h))
		}
		return h * sum / 3
	case "gauss":
		for i := 0; i < n; i++ {
			mid := a + (float64(i)+0.5)*h
			for _, p := range rule {
				add(p.Y * f(mid+h*p.X/2))
			}
		}
		return h * sum / 2
	}
	return math.NaN()
}

func Integrate(in IntegrationInput) (IntegrationResult, error) {
	if strings.TrimSpace(in.Expression) == "" {
		in.Expression = DefaultIntegrand
	}
	out := IntegrationResult{Input: in, GaussNodes: []Point{}}
	out.HasReference = strings.Join(strings.Fields(in.Expression), "") == DefaultIntegrand
	if !finite(in.A) || !finite(in.B) || in.B-in.A < 1e-9 || math.Abs(in.A) > 1e6 || math.Abs(in.B) > 1e6 {
		return out, fmt.Errorf("нужны конечные границы −1 000 000 ≤ a < b ≤ 1 000 000; длина интервала не меньше 10⁻⁹")
	}
	if out.HasReference && (in.A < -.99 || in.B > .99) {
		return out, fmt.Errorf("для функции варианта 9 границы должны удовлетворять −0,99 ≤ a < b ≤ 0,99")
	}
	formula, err := compileExpression(in.Expression)
	if err != nil {
		return out, err
	}
	var evaluationError error
	f := func(x float64) float64 {
		if evaluationError != nil {
			return math.NaN()
		}
		v := formula.eval(x).v
		if !finite(v) {
			evaluationError = fmt.Errorf("функция не определена или переполняется при x = %g; проверьте формулу и границы", x)
		}
		return v
	}
	// Include boundaries even for methods with interior-only nodes.
	for i := 0; i <= 400; i++ {
		x := in.A + (in.B-in.A)*float64(i)/400
		v := f(x)
		if evaluationError != nil {
			return out, evaluationError
		}
		out.Curve = append(out.Curve, Point{x, v})
	}
	if !finite(in.Tolerance) || in.Tolerance < 1e-12 || in.Tolerance > .1 {
		return out, fmt.Errorf("относительная точность должна быть от 10⁻¹² до 0,1")
	}
	if in.N < 1 || in.N > 4096 {
		return out, fmt.Errorf("начальное число интервалов должно быть от 1 до 4096")
	}
	if in.Order < 2 || in.Order > 16 {
		return out, fmt.Errorf("порядок Гаусса должен быть от 2 до 16")
	}
	p := 2
	var rule []Point
	switch in.Method {
	case "midpoint", "trapezoid":
	case "simpson":
		if in.N%2 != 0 {
			return out, fmt.Errorf("для Симпсона начальное число интервалов должно быть чётным")
		}
		p = 4
	case "gauss":
		rule = GaussRule(in.Order)
		out.GaussNodes = rule
	default:
		return out, fmt.Errorf("неизвестный метод интегрирования")
	}
	if out.HasReference {
		out.Reference = ReferencePrimitive(in.B) - ReferencePrimitive(in.A)
	}
	n := in.N
	if in.Method == "gauss" {
		n = 1
	}
	previous := quadrature(f, in.A, in.B, n, in.Method, rule)
	if evaluationError != nil {
		return out, evaluationError
	}
	if !finite(previous) {
		return out, fmt.Errorf("переполнение при интегрировании")
	}
	out.Iterations = append(out.Iterations, Iteration{N: n, H: (in.B - in.A) / float64(n), Value: previous})
	maxIntervals := 1 << 20
	if !out.HasReference {
		maxIntervals = 1 << 16
	}
	for n <= maxIntervals/2 {
		n *= 2
		value := quadrature(f, in.A, in.B, n, in.Method, rule)
		if evaluationError != nil {
			return out, evaluationError
		}
		if !finite(value) {
			return out, fmt.Errorf("переполнение при интегрировании")
		}
		estimate := math.Abs(value-previous) / float64((int64(1)<<p)-1)
		// For Gauss use the undivided difference: the asymptotic 2^(2m)-1
		// divisor is overly optimistic before reaching the asymptotic regime.
		if in.Method == "gauss" {
			estimate = math.Abs(value - previous)
		}
		out.AbsoluteCriterion = math.Abs(value) < 1e-12
		scale := math.Abs(value)
		if out.AbsoluteCriterion {
			scale = 1
		}
		relative := estimate / scale
		out.Iterations = append(out.Iterations, Iteration{n, (in.B - in.A) / float64(n), value, &estimate, &relative})
		out.Value = value
		out.N = n
		out.Estimate = estimate
		if out.HasReference {
			out.AbsoluteError = math.Abs(value - out.Reference)
			out.RelativeError = out.AbsoluteError / math.Abs(out.Reference)
		}
		// The known function permits an independent verification of the target.
		if relative <= in.Tolerance && (!out.HasReference || out.RelativeError <= in.Tolerance) {
			out.Converged = true
			break
		}
		previous = value
	}
	out.Message = "Точность достигнута и проверена по независимому степенному ряду."
	if !out.HasReference {
		out.Message = "Критерий точности выполнен по оценке последовательных приближений; точный интеграл неизвестен."
	}
	if out.AbsoluteCriterion {
		out.Message = "Интеграл близок к нулю: ε используется как абсолютная точность. Критерий выполнен по оценке метода."
	}
	if !out.Converged {
		out.Message = fmt.Sprintf("Достигнут предел %d интервалов; заданная точность не достигнута.", maxIntervals)
	}
	return out, nil
}
