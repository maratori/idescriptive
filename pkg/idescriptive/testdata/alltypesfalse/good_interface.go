package alltypesfalse

type Empty interface{}

type NoParams interface {
	NoReturn()
	SingleReturn() error
	DoubleReturn() (x int, err error)
}

type SingleParam interface {
	NoReturn(a int)
	SingleReturn(b bool) (c int)
	DoubleReturn(Empty) (d int, err error)
}
