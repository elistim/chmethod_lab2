// Package numerics implements the methods from tasks 6.3–6.5, variant 9.
package numerics

import (
	"fmt"
	"math"
	"sort"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type InterpolationInput struct {
	Nodes       []Point   `json:"nodes"`
	Grid        []float64 `json:"grid"`
	Degree      int       `json:"degree"`
	Reference   bool      `json:"reference"`
	Expression  string    `json:"expression"`
	UseFunction bool      `json:"useFunction"`
}
type InterpolationRow struct {
	X           float64  `json:"x"`
	Value       float64  `json:"value"`
	First       float64  `json:"first"`
	Second      float64  `json:"second"`
	Exact       *float64 `json:"exact"`
	ExactFirst  *float64 `json:"exactFirst"`
	ExactSecond *float64 `json:"exactSecond"`
	Error       *float64 `json:"error"`
	ErrorFirst  *float64 `json:"errorFirst"`
	ErrorSecond *float64 `json:"errorSecond"`
	Start       int      `json:"start"`
	End         int      `json:"end"`
}
type InterpolationResult struct {
	Rows       []InterpolationRow `json:"rows"`
	Curve      []InterpolationRow `json:"curve"`
	Nodes      []Point            `json:"nodes"`
	Degree     int                `json:"degree"`
	Reference  bool               `json:"reference"`
	MaxError   [3]float64         `json:"maxError"`
	Expression string             `json:"expression"`
}

func finite(v float64) bool      { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func Function(x float64) float64 { return math.Sqrt(x) + 1 }
func Derivatives(x float64) (float64, float64) {
	return 1 / (2 * math.Sqrt(x)), -1 / (4 * x * math.Sqrt(x))
}

func DefaultInterpolation() InterpolationInput {
	in := InterpolationInput{Degree: 3, Reference: true, Expression: DefaultFunction, UseFunction: true}
	for i := 0; i <= 10; i++ {
		x := 1 + float64(i)/10
		in.Nodes = append(in.Nodes, Point{x, Function(x)})
	}
	for i := 0; i <= 20; i++ {
		in.Grid = append(in.Grid, 1+float64(i)/20)
	}
	return in
}

// Newton evaluates a polynomial and its analytic first and second derivatives.
// Affine scaling keeps divided differences away from very small powers of h.
func Newton(nodes []Point, x float64) (value, first, second float64) {
	n := len(nodes)
	origin := nodes[0].X
	scale := nodes[n-1].X - origin
	if n == 1 {
		return nodes[0].Y, 0, 0
	}
	z := make([]float64, n)
	c := make([]float64, n)
	for i, p := range nodes {
		z[i] = (p.X - origin) / scale
		c[i] = p.Y
	}
	for k := 1; k < n; k++ {
		for i := n - 1; i >= k; i-- {
			c[i] = (c[i] - c[i-1]) / (z[i] - z[i-k])
		}
	}
	t := (x - origin) / scale
	value = c[n-1]
	for i := n - 2; i >= 0; i-- {
		second = second*(t-z[i]) + 2*first
		first = first*(t-z[i]) + value
		value = value*(t-z[i]) + c[i]
	}
	return value, first / scale, second / (scale * scale)
}

// Select degree+1 nearest nodes, retaining their natural order. In one
// dimension these form a contiguous window. Ties choose the left node.
func window(nodes []Point, x float64, degree int) int {
	r := sort.Search(len(nodes), func(i int) bool { return nodes[i].X >= x })
	l := r - 1
	for i := 0; i <= degree; i++ {
		if l < 0 {
			r++
		} else if r >= len(nodes) || x-nodes[l].X <= nodes[r].X-x {
			l--
		} else {
			r++
		}
	}
	return l + 1
}

func Interpolate(in InterpolationInput) (InterpolationResult, error) {
	if in.Expression == "" {
		in.Expression = DefaultFunction
	}
	out := InterpolationResult{Degree: in.Degree, Reference: in.Reference, Expression: in.Expression}
	var formula *expression
	if in.Reference || in.UseFunction {
		var err error
		formula, err = compileExpression(in.Expression)
		if err != nil {
			return out, err
		}
	}
	in.Nodes = append([]Point(nil), in.Nodes...)
	if len(in.Nodes) < 2 || len(in.Nodes) > 201 {
		return out, fmt.Errorf("нужно от 2 до 201 исходного узла")
	}
	if in.Degree < 1 || in.Degree > 20 || in.Degree >= len(in.Nodes) {
		return out, fmt.Errorf("степень должна быть от 1 до min(20, число узлов − 1)")
	}
	if len(in.Grid) < 1 || len(in.Grid) > 2001 {
		return out, fmt.Errorf("новая сетка должна содержать от 1 до 2001 узла")
	}
	for i, p := range in.Nodes {
		if in.UseFunction {
			p.Y = formula.eval(p.X).v
			in.Nodes[i].Y = p.Y
		}
		if !finite(p.X) || !finite(p.Y) || math.Abs(p.X) > 1e6 || math.Abs(p.Y) > 1e12 {
			return out, fmt.Errorf("некорректные данные узла %d", i)
		}
		if i > 0 && p.X-in.Nodes[i-1].X < 1e-9 {
			return out, fmt.Errorf("исходные x должны строго возрастать; минимальное расстояние 10⁻⁹")
		}
		if in.Reference {
			f := formula.eval(p.X).v
			if !finite(f) || math.Abs(p.Y-f) > 1e-12*math.Max(1, math.Abs(f)) {
				return out, fmt.Errorf("значения таблицы не совпадают с формулой; включите вычисление y по формуле или отключите сравнение")
			}
		}
	}
	a, b := in.Nodes[0].X, in.Nodes[len(in.Nodes)-1].X
	for _, x := range in.Grid {
		if !finite(x) || x < a || x > b {
			return out, fmt.Errorf("новая сетка должна лежать внутри [%g; %g]", a, b)
		}
	}
	eval := func(x float64) (InterpolationRow, error) {
		s := window(in.Nodes, x, in.Degree)
		v, d1, d2 := Newton(in.Nodes[s:s+in.Degree+1], x)
		row := InterpolationRow{X: x, Value: v, First: d1, Second: d2, Start: s, End: s + in.Degree}
		if !finite(v) || !finite(d1) || !finite(d2) {
			return row, fmt.Errorf("переполнение: уменьшите степень или измените узлы")
		}
		if in.Reference {
			analytic := formula.eval(x)
			f, f1, f2 := analytic.v, analytic.d, analytic.dd
			if !finite(f) || !finite(f1) || !finite(f2) {
				return row, fmt.Errorf("функция или её производные не определены при x = %g; измените интервал или отключите аналитическое сравнение", x)
			}
			e, e1, e2 := math.Abs(v-f), math.Abs(d1-f1), math.Abs(d2-f2)
			row.Exact = &f
			row.ExactFirst = &f1
			row.ExactSecond = &f2
			row.Error = &e
			row.ErrorFirst = &e1
			row.ErrorSecond = &e2
		}
		return row, nil
	}
	for _, x := range in.Grid {
		r, e := eval(x)
		if e != nil {
			return out, e
		}
		out.Rows = append(out.Rows, r)
		if in.Reference {
			out.MaxError[0] = math.Max(out.MaxError[0], *r.Error)
			out.MaxError[1] = math.Max(out.MaxError[1], *r.ErrorFirst)
			out.MaxError[2] = math.Max(out.MaxError[2], *r.ErrorSecond)
		}
	}
	for i := 0; i <= 400; i++ {
		r, e := eval(a + (b-a)*float64(i)/400)
		if e != nil {
			return out, e
		}
		out.Curve = append(out.Curve, r)
	}
	out.Nodes = append([]Point(nil), in.Nodes...)
	return out, nil
}
