package strict_true

type BadSingleParam interface {
	NoReturn(int)                          // want `missing name for incoming parameter int in method NoReturn\(int\)`
	SingleReturn(bool) (x int)             // want `missing name for incoming parameter bool in method SingleReturn\(bool\) \(x int\)`
	DoubleReturn(Empty) (y int, err error) // want `missing name for incoming parameter Empty in method DoubleReturn\(Empty\) \(y int, err error\)`
}
