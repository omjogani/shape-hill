package hillchart

import "math"

const (
	width    = 900.0
	left     = 60.0
	right    = 840.0
	baseline = 300.0
	peak     = 60.0
	rowStep  = 26.0
)

// x maps a 0-100 position onto the baseline.
func x(position int16) float64 {
	return left + clamp(position)/100*(right-left)
}

// y is a raised cosine: flat at both feet and flat over the summit, which is what
// makes a dot's height read as certainty rather than distance travelled.
func y(position int16) float64 {
	t := clamp(position) / 100
	return baseline - (baseline-peak)*(1-math.Cos(2*math.Pi*t))/2
}

func clamp(position int16) float64 {
	switch {
	case position < 0:
		return 0
	case position > 100:
		return 100
	default:
		return float64(position)
	}
}
