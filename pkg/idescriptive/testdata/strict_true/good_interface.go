package strict_true

type Empty interface{}

type NoParams interface {
	NoReturn()
	SingleReturn() (x int)
	DoubleReturn() (y int, err error)
}

type SingleParam interface {
	NoReturn(a int)
	SingleReturn(b bool) (c int)
	DoubleReturn(e Empty) (d int, err error)
}
