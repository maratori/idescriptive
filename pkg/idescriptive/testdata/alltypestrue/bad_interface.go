package alltypestrue

type BadSingleParam interface {
	NoReturn(int)                          // want `missing input parameter name`
	SingleReturn(bool) (x int)             // want `missing input parameter name`
	DoubleReturn(Empty) (y int, err error) // want `missing input parameter name`
}
