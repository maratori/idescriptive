package strict_false

type BadSingleParam interface {
	NoReturn(int)              // want `missing name for incoming parameter int in method NoReturn\(int\)`
	SingleReturn(bool) (x int) // want `missing name for incoming parameter bool in method SingleReturn\(bool\) \(x int\)`
	DoubleReturn(Empty) (x int, err error)
}
