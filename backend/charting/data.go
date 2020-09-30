package charting

import "time"

type Data struct {
	X            []time.Time
	Y            []interface{}
	Label        []string
	uniqueLabels map[string]bool
}

func NewData() *Data {
	return &Data{uniqueLabels: make(map[string]bool)}
}

func (ld *Data) Append(x time.Time, y interface{}, label string) {
	ld.X = append(ld.X, x)
	ld.Y = append(ld.Y, y)
	ld.Label = append(ld.Label, label)
	ld.uniqueLabels[label] = true
}

func (ld *Data) Clone() *Data {
	return &Data{
		X:            append([]time.Time{}, ld.X...),
		Y:            append([]interface{}{}, ld.Y...),
		Label:        append([]string{}, ld.Label...),
		uniqueLabels: map[string]bool{}}
}

func (ld *Data) SplitByLabels(zeroValue interface{}) map[string]*Data {
	ret := map[string]*Data{}
	for key := range ld.uniqueLabels {
		clone := ld.Clone()
		ret[key] = clone
		for pos, label := range clone.Label {
			if label != key {
				clone.Y[pos] = zeroValue
				clone.Label[pos] = ""
			}
		}
	}
	return ret
}
