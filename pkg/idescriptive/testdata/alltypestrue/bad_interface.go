package alltypestrue

type BadSingleParam interface {
	NoReturn(int)                          // want `missing incoming parameter name`
	SingleReturn(bool) (x int)             // want `missing incoming parameter name`
	DoubleReturn(Empty) (y int, err error) // want `missing incoming parameter name`
}
