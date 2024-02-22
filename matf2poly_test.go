package golsv

import (
	"testing"
)

func TestMatF2PolyAdd(t *testing.T) {
	tests := []struct {
		M1, M2, want	MatF2Poly
	}{
		{ MatF2Poly{}, MatF2Poly{}, MatF2Poly{} },
		{ MatF2PolyIdentity, MatF2PolyIdentity, MatF2Poly{} },
		{ MatF2PolyIdentity, MatF2Poly{}, MatF2PolyIdentity },
		{ NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"),
			NewMatF2PolyFromString("[0 0 0 0 0 0 1 0 0]"),
			NewMatF2PolyFromString("[11 0 01 01 11 01 1 01 11]") },
	}
	for _, test := range tests {
		if got := test.M1.Add(test.M2); !got.Equal(test.want) {
			t.Errorf("MatF2Poly.Add() M1=%v M2=%v got=%v want=%v", test.M1, test.M2, got, test.want)
		}
	}
}

func TestMatF2PolyEqual(t *testing.T) {
	tests := []struct {
		M1, M2	MatF2Poly
		want	bool
	}{
		{ MatF2Poly{}, MatF2Poly{}, true },
		{ MatF2PolyIdentity, MatF2PolyIdentity, true },
		{ MatF2Poly{}, MatF2PolyIdentity, false },
	}
	for _, test := range tests {
		if got := test.M1.Equal(test.M2); got != test.want {
			t.Errorf("MatF2Poly.Equal() M1=%v M2=%v got=%v want=%v", test.M1, test.M2, got, test.want)
		}
	}
}

func TestMatF2PolyMul(t *testing.T) {
	tests := []struct {
		M1, M2, want	MatF2Poly
	}{
		{ MatF2Poly{}, MatF2Poly{}, MatF2Poly{} },
		{ MatF2PolyIdentity, MatF2PolyIdentity, MatF2PolyIdentity },
		{ NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"), MatF2PolyIdentity,
			NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]") },
		{ NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"),
			NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"),
			NewMatF2PolyFromString("[101 001 0 0 1 001 001 0 1]")},
	}
	for _, test := range tests {
		if got := test.M1.Mul(test.M2); !got.Equal(test.want) {
			t.Errorf("MatF2Poly.Mul() M1=%v M2=%v got=%v want=%v", test.M1, test.M2, got, test.want)
		}
	}
}

func TestMatF2PolyPow(t *testing.T) {
	tests := []struct {
		M	MatF2Poly
		n	int
		want	MatF2Poly
	}{
		{ MatF2Poly{}, 0, MatF2PolyIdentity },
		{ MatF2PolyIdentity, 0, MatF2PolyIdentity },
		{ MatF2PolyIdentity, 1, MatF2PolyIdentity },
		{ MatF2PolyIdentity, 2, MatF2PolyIdentity },
		{ NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"), 2,
			NewMatF2PolyFromString("[101 001 0 0 1 001 001 0 1]")},
	}
	for _, test := range tests {
		if got := test.M.Pow(test.n); !got.Equal(test.want) {
			t.Errorf("MatF2Poly.Pow() M=%v n=%v got=%v want=%v", test.M, test.n, got, test.want)
		}
	}
}

func TestMatF2PolyScale(t *testing.T) {
	tests := []struct {
		M	MatF2Poly
		c	F2Polynomial
		want	MatF2Poly
	}{
		{ MatF2Poly{}, F2PolynomialZero, MatF2Poly{} },
		{ MatF2PolyIdentity, F2PolynomialZero, MatF2Poly{} },
		{ MatF2PolyIdentity, F2PolynomialOne, MatF2PolyIdentity },
		{ NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"), F2PolynomialOne,
			NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]") },
		{
			NewMatF2PolyFromString("[11 0 01 01 11 01 0 01 11]"),
			NewF2Polynomial("01"),
			NewMatF2PolyFromString("[011 0 001 001 011 001 0 001 011]") },
	}
	for _, test := range tests {
		if got := test.M.Scale(test.c); !got.Equal(test.want) {
			t.Errorf("MatF2Poly.Scale() M=%v c=%v got=%v want=%v", test.M, test.c, got, test.want)
		}
	}
}

func TestMatF2PolyTrace(t *testing.T) {
	tests := []struct {
		M	MatF2Poly
		want	F2Polynomial
	}{
		{ MatF2Poly{}, F2PolynomialZero },
		{ MatF2PolyIdentity, F2PolynomialOne },
		{ NewMatF2Poly(
			F2PolynomialOne, F2PolynomialZero, F2PolynomialOne,
			F2PolynomialOne, F2PolynomialZero, F2PolynomialOne,
			F2PolynomialZero, F2PolynomialZero, F2PolynomialOne,),
			F2PolynomialZero },
	}
	for _, test := range tests {
		if got := test.M.Trace(); got != test.want {
			t.Errorf("MatF2Poly.Trace() M=%v got=%v want=%v", test.M, got, test.want)
		}
	}
}

func TestProjMatF2PolyEqual(t *testing.T) {
	tests := []struct {
		M1, M2	ProjMatF2Poly
		want	bool
	}{
		{ ProjMatF2Poly{}, ProjMatF2Poly{}, true },
		{ ProjMatF2PolyIdentity, ProjMatF2PolyIdentity, true },
		{ ProjMatF2Poly{}, ProjMatF2PolyIdentity, false },
		{
			NewProjMatF2PolyFromString("[11 01 0 0 11 01 01 1 01]"),
			NewProjMatF2PolyFromString("[1 01 0 0 1 01 01 1 0]"),
			false },
		{
			NewProjMatF2PolyFromString("[1 01 0 0 1 01 01 1 0]"),
			NewProjMatF2PolyFromString("[11 011 0 0 11 011 011 11 0]"),
			true },
	}
	for n, test := range tests {
		if got := test.M1.Equal(test.M2); got != test.want {
			t.Errorf("ProjMatF2Poly.Equal() n=%d M1=%v M2=%v got=%v want=%v", n, test.M1, test.M2, got, test.want)
		}
	}
}

func TestProjMatF2PolyReduceModf(t *testing.T) {
	tests := []struct {
		M		ProjMatF2Poly
		f		F2Polynomial
		want	ProjMatF2Poly
	}{
		{ ProjMatF2Poly{}, F2PolynomialOne, ProjMatF2Poly{} },
		{ ProjMatF2PolyIdentity, F2PolynomialOne, ProjMatF2Poly{} },
		{
			NewProjMatF2PolyFromString("[101 01 0 0 11 01 01 1 01]"),
			NewF2Polynomial("11"),
			NewProjMatF2PolyFromString("[0 1 0 0 0 1 1 1 1]") },
	}
	for _, test := range tests {
		if got := test.M.ReduceModf(test.f); !got.Equal(test.want) {
			t.Errorf("ProjMatF2Poly.ReduceModf() M=%v f=%v got=%v want=%v", test.M, test.f, got, test.want)
		}
	}
}
