package measurement

type Measurement interface {
	Add(Number) Number
	Get() Number
	Reset()
	Update(func(Number) Number)
}
