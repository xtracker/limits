package measurement

func NewAverageMeasurement(window, warmupWindow int) *ExpAvgMeasurement {
	return &ExpAvgMeasurement{
		window:       window,
		warmupWindow: warmupWindow,
		sum:          0.0,
		value:        0.0,
		count:        0,
	}
}

type ExpAvgMeasurement struct {
	value        float64
	sum          float64
	window       int
	warmupWindow int
	count        int
}

func (m *ExpAvgMeasurement) Add(sample Number) Number {
	if m.count < m.warmupWindow {
		m.count = m.count + 1
		m.sum = m.sum + sample.Float64()
		m.value = m.sum / float64(m.count)
	} else {
		factor := 2.0 / float64(m.window+1)
		m.value = m.value*(1.0-factor) + sample.Float64()*factor
	}

	return Float64Number(m.value)
}

func (m *ExpAvgMeasurement) Get() Number {
	return Float64Number(m.value)
}

func (m *ExpAvgMeasurement) Reset() {
	m.value = 0.0
	m.count = 0
	m.sum = 0
}

func (m *ExpAvgMeasurement) Update(op func(Number) Number) {
	m.value = op(Float64Number(m.value)).Float64()
}
