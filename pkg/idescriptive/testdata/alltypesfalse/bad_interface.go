package alltypesfalse

type BadSingleParam interface {
	NoReturn(int)              // want `missing incoming parameter name`
	SingleReturn(bool) (x int) // want `missing incoming parameter name`
	DoubleReturn(Empty) (x int, err error)
}
