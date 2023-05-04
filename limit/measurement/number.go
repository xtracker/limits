package measurement

type Number interface {
	Float64() float64
	Int64() int64
	Int() int
}

type Float64Number float64

func (f Float64Number) Float64() float64 {
	return float64(f)
}

func (f Float64Number) Int64() int64 {
	return int64(f)
}

func (f Float64Number) Int() int {
	return int(f)
}

type Int64Number int64

func (n Int64Number) Float64() float64 {
	return float64(n)
}

func (n Int64Number) Int64() int64 {
	return int64(n)
}

func (n Int64Number) Int() int {
	return int(n)
}
