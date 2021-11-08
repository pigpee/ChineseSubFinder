package sub_timeline_fixer

import "testing"

func TestCalculateCurveCorrelation(t *testing.T) {
	type args struct {
		s1 []float64
		s2 []float64
		n  int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "00", args: args{
			s1: []float64{0.309016989, 0.587785244, 0.809016985, 0.95105651, 1, 0.951056526,
				0.809017016, 0.587785287, 0.30901704, 5.35898e-08, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0},
			s2: []float64{0.343282816, 0.686491368, 0.874624132, 0.99459642, 1.008448609,
				1.014252458, 0.884609221, 0.677632906, 0.378334666, 0.077878732,
				0.050711886, 0.066417083, 0.088759401, 0.005440732, 0.04225661,
				0.035349939, 0.0631196, 0.007566056, 0.053183895, 0.073143706,
				0.080285063, 0.030110227, 0.044781145, 0.01875573, 0.08373928,
				0.04550342, 0.038880858, 0.040611891, 0.046116826, 0.087670453},
			n: 30,
		},
			want: 0.99743484574875},
		{name: "01", args: args{
			s1: []float64{0.309016989, 0.587785244, 0.809016985, 0.95105651, 1, 0.951056526,
				0.809017016, 0.587785287, 0.30901704, 5.35898e-08, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0},
			s2: []float64{0.309016989, 0.587785244, 0.809016985, 0.95105651, 1, 0.951056526,
				0.809017016, 0.587785287, 0.30901704, 5.35898e-08, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0},
			n: 30,
		},
			want: 1},
		{name: "02", args: args{
			s1: []float64{0.309016989, 0.587785244, 0.809016985, 0.95105651, 1, 0.951056526,
				0.809017016, 0.587785287, 0.30901704, 5.35898e-08, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0},
			s2: []float64{-0.309016989, -0.587785244, -0.809016985, -0.95105651, -1, -0.951056526,
				-0.809017016, -0.587785287, -0.30901704, -5.35898e-08, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0},
			n: 30,
		},
			want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateCurveCorrelation(tt.args.s1, tt.args.s2, tt.args.n); got != tt.want {
				t.Errorf("CalculateCurveCorrelation() = %v, want %v", got, tt.want)
			}
		})
	}
}
