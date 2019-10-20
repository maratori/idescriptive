package alltypesfalse

type BadSingleParam interface {
	NoReturn(int)              // want `missing input parameter name`
	SingleReturn(bool) (x int) // want `missing input parameter name`
	DoubleReturn(Empty) (x int, err error)
}
