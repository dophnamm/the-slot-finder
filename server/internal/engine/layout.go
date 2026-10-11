package engine

import (
	"math"
	"sort"
)

// epsilon absorbs floating-point noise when comparing hours.
const epsilon = 1e-6

// layoutMode says whether a layout may place less than the minimum in a week.
type layoutMode int

const (
	// seatLayout never places less than MinHoursPerWeek in a week.
	seatLayout layoutMode = iota
	// remainderLayout may place a demand's last remainder, below the minimum, in one week.
	remainderLayout
)

// flatLayout spreads need over the weeks as evenly as the free hours allow (the flat
// profile). It returns one value per week; their sum is what the person can cover.
func flatLayout(need float64, free []float64, minHours, step float64, mode layoutMode) []float64 {
	out := make([]float64, len(free))
	if need <= epsilon || len(free) == 0 {
		return out
	}
	if need < minHours-epsilon {
		if mode == remainderLayout {
			placeInOneWeek(need, free, out, step)
		}
		return out
	}

	eligible := make([]int, 0, len(free))
	for i, f := range free {
		if f >= minHours-epsilon && f > epsilon {
			eligible = append(eligible, i)
		}
	}
	if len(eligible) == 0 {
		return out
	}
	weeks := len(eligible)
	if minHours > epsilon {
		weeks = min(len(eligible), max(1, int(math.Floor(need/minHours+epsilon))))
	}
	chosen := evenlySpaced(eligible, weeks)
	fill := waterFill(need, pick(free, chosen))
	if sum(fill) < need-epsilon && len(chosen) < len(eligible) {
		chosen = eligible
		fill = waterFill(need, pick(free, chosen))
	}
	chosen, fill = dropBelowMinimum(need, free, chosen, fill, minHours)
	for i, week := range chosen {
		out[week] = fill[i]
	}
	return quantize(out, free, step)
}

// placeInOneWeek puts all of need into the first week that has room for it, or nothing.
func placeInOneWeek(need float64, free, out []float64, step float64) {
	rounded := roundToStep(need, step)
	if rounded <= epsilon {
		return
	}
	for i, f := range free {
		if f >= rounded-epsilon {
			out[i] = rounded
			return
		}
	}
}

// dropBelowMinimum removes the week with the smallest share while any share is below the
// minimum, filling again over the weeks left, so no week ends up below the minimum.
func dropBelowMinimum(need float64, free []float64, chosen []int, fill []float64, minHours float64) ([]int, []float64) {
	for len(chosen) > 0 {
		smallest := -1
		for i, hours := range fill {
			if hours < minHours-epsilon && (smallest < 0 || hours < fill[smallest]) {
				smallest = i
			}
		}
		if smallest < 0 {
			return chosen, fill
		}
		chosen = append(append([]int{}, chosen[:smallest]...), chosen[smallest+1:]...)
		fill = waterFill(need, pick(free, chosen))
	}
	return chosen, fill
}

// waterFill shares need over the caps as evenly as possible: every value is
// min(cap, level) for one common level, and the values sum to need when the caps allow.
func waterFill(need float64, caps []float64) []float64 {
	out := make([]float64, len(caps))
	order := make([]int, len(caps))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return caps[order[a]] < caps[order[b]] })
	left := need
	for n, i := range order {
		share := left / float64(len(order)-n)
		out[i] = math.Max(0, math.Min(caps[i], share))
		left -= out[i]
	}
	return out
}

// quantize rounds values to the step while keeping their rounded total and the caps.
// The units lost by rounding down go to the largest fractions first, earliest week on ties.
func quantize(values, caps []float64, step float64) []float64 {
	if step <= epsilon {
		return values
	}
	units := make([]float64, len(values))
	capUnits := make([]float64, len(values))
	total := math.Round(sum(values) / step)
	placed := 0.0
	for i, v := range values {
		capUnits[i] = math.Floor(caps[i]/step + epsilon)
		units[i] = math.Min(math.Floor(v/step+epsilon), capUnits[i])
		placed += units[i]
	}
	order := make([]int, len(values))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return fraction(values[order[a]], step) > fraction(values[order[b]], step)
	})
	for placed < total-epsilon {
		progressed := false
		for _, i := range order {
			if placed >= total-epsilon {
				break
			}
			if values[i] > epsilon && units[i]+1 <= capUnits[i] {
				units[i]++
				placed++
				progressed = true
			}
		}
		if !progressed {
			break
		}
	}
	out := make([]float64, len(values))
	for i, u := range units {
		out[i] = u * step
	}
	return out
}

func fraction(value, step float64) float64 {
	units := value / step
	return units - math.Floor(units+epsilon)
}

// roundToStep rounds down to the step; a zero step keeps the value.
func roundToStep(value, step float64) float64 {
	if step <= epsilon {
		return value
	}
	return math.Floor(value/step+epsilon) * step
}

// evenlySpaced picks k of the items spread from the first to the last.
func evenlySpaced(items []int, k int) []int {
	if k >= len(items) {
		return items
	}
	if k == 1 {
		return []int{items[0]}
	}
	out := make([]int, 0, k)
	for i := range k {
		out = append(out, items[int(math.Round(float64(i)*float64(len(items)-1)/float64(k-1)))])
	}
	return out
}

func pick(values []float64, indexes []int) []float64 {
	out := make([]float64, len(indexes))
	for i, index := range indexes {
		out[i] = values[index]
	}
	return out
}

func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}
