package numerics

import (
	"math"
	"testing"
)

func near(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("got %.16g, want %.16g (tol %g)", got, want, tolerance)
	}
}

func TestNewtonPolynomialAndDerivatives(t *testing.T) {
	f := func(x float64) float64 { return 2*x*x*x - 3*x*x + 4*x - 7 }
	nodes := []Point{{-2, f(-2)}, {-.4, f(-.4)}, {.8, f(.8)}, {3, f(3)}}
	for _, x := range []float64{-2, -1.1, 0, .8, 2.2, 3} {
		v, d1, d2 := Newton(nodes, x)
		near(t, v, f(x), 1e-12)
		near(t, d1, 6*x*x-6*x+4, 1e-12)
		near(t, d2, 12*x-6, 1e-12)
	}
}
func TestVariantInterpolation(t *testing.T) {
	in := DefaultInterpolation()
	out, err := Interpolate(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Rows) != 21 || len(out.Nodes) != 11 {
		t.Fatal("wrong variant grid")
	}
	near(t, out.Nodes[0].Y, 2, 0)
	near(t, out.Nodes[10].Y, 1+math.Sqrt(2), 0)
	for i, row := range out.Rows {
		if i%2 == 0 {
			near(t, row.Value, *row.Exact, 2e-15)
		}
		if row.Start < 0 || row.End >= 11 || row.End-row.Start != 3 {
			t.Fatal("bad local window")
		}
	}
	// Endpoint second derivatives of the cubic have O(h²) error.
	if out.MaxError[0] > 4e-6 || out.MaxError[1] > 2e-4 || out.MaxError[2] > .007 {
		t.Fatalf("unexpected errors: %v", out.MaxError)
	}
	for _, degree := range []int{1, 2, 10} {
		in.Degree = degree
		if _, err := Interpolate(in); err != nil {
			t.Fatal(err)
		}
	}
}
func TestCustomTableAndValidation(t *testing.T) {
	in := InterpolationInput{Nodes: []Point{{0, 1}, {1, 3}, {2, 5}}, Grid: []float64{0, .5, 2}, Degree: 1}
	out, err := Interpolate(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range out.Rows {
		near(t, r.Value, 1+2*r.X, 1e-14)
		near(t, r.First, 2, 1e-14)
		near(t, r.Second, 0, 0)
		if r.Error != nil {
			t.Fatal("unknown function must not have exact error")
		}
	}
	in.Reference = true
	if _, err := Interpolate(in); err == nil {
		t.Fatal("invalid reference accepted")
	}
	in.Reference = false
	in.Nodes[1].X = 0
	if _, err := Interpolate(in); err == nil {
		t.Fatal("duplicate nodes accepted")
	}
	in = DefaultInterpolation()
	in.Grid = []float64{2.1}
	if _, err := Interpolate(in); err == nil {
		t.Fatal("extrapolation accepted")
	}
}
func TestGaussPolynomialExactness(t *testing.T) {
	for order := 2; order <= 16; order++ {
		rule := GaussRule(order)
		for degree := 0; degree < 2*order; degree++ {
			got := quadrature(func(x float64) float64 { return math.Pow(x, float64(degree)) }, -1, 1, 1, "gauss", rule)
			want := 0.0
			if degree%2 == 0 {
				want = 2 / float64(degree+1)
			}
			near(t, got, want, 6e-14)
		}
	}
}
func TestQuadratureExactness(t *testing.T) {
	for _, method := range []string{"trapezoid", "midpoint", "simpson"} {
		near(t, quadrature(func(x float64) float64 { return 3*x + 2 }, -.7, 2.3, 8, method, nil), 13.2, 1e-13)
	}
	near(t, quadrature(func(x float64) float64 { return x * x * x }, 0, 2, 8, "simpson", nil), 4, 1e-14)
}
func TestVariantIntegration(t *testing.T) {
	for _, method := range []string{"simpson", "trapezoid", "midpoint", "gauss"} {
		out, err := Integrate(IntegrationInput{A: -.75, B: .75, N: 8, Tolerance: 1e-8, Method: method, Order: 8})
		if err != nil {
			t.Fatal(err)
		}
		if !out.Converged || out.RelativeError > 1e-8 {
			t.Fatalf("%s failed: %+v", method, out)
		}
		// Independent fine composite Simpson quadrature verifies the series.
		near(t, out.Reference, quadrature(Integrand, -.75, .75, 65536, "simpson", nil), 2e-14)
		if out.Iterations[0].N != 8 && method != "gauss" {
			t.Fatal("wrong initial n")
		}
		t.Logf("%s: I=%.15g, n=%d, relative error=%.3g", method, out.Value, out.N, out.RelativeError)
	}
}
func TestIntegrationDomainAndAccuracy(t *testing.T) {
	in := IntegrationInput{A: -.99, B: .98, N: 8, Tolerance: 1e-10, Method: "gauss", Order: 8}
	out, err := Integrate(in)
	if err != nil || !out.Converged {
		t.Fatalf("near boundary: %v", err)
	}
	near(t, out.Reference, quadrature(Integrand, in.A, in.B, 65536, "simpson", nil), 1e-11)
	in.Method = "simpson"
	in.N = 3
	if _, err := Integrate(in); err == nil {
		t.Fatal("odd Simpson n accepted")
	}
	in.N = 8
	in.A = -1
	if _, err := Integrate(in); err == nil {
		t.Fatal("singularity accepted")
	}
	in.A = -.75
	in.Tolerance = 0
	if _, err := Integrate(in); err == nil {
		t.Fatal("zero tolerance accepted")
	}
}
