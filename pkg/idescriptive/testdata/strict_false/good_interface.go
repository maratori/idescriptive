package strict_false

type Empty interface{}

type NoParams interface {
	NoReturn()
	SingleReturn() error
	DoubleReturn() (x int, err error)
}

type OneParam interface {
	NoReturn(int)
	SingleReturn(bool) (c int)
	DoubleReturn(Empty) (d int, err error)
}

type TwoParams interface {
	NoReturn(a, b int)
	NoReturn2(a string, b int)
	SingleReturn(b, c bool) (d int)
	SingleReturn2(b bool, c string) (d int)
	DoubleReturn(Empty) (d int, err error)
}
