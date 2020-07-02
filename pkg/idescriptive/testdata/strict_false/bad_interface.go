package strict_false

type BadTwoParams interface {
	NoReturn(int, int)                 // want `missing name for incoming parameter int in method NoReturn\(int, int\)` `missing name for incoming parameter int in method NoReturn\(int, int\)`
	SingleReturn(bool, string) (x int) // want `missing name for incoming parameter bool in method SingleReturn\(bool, string\) \(x int\)` `missing name for incoming parameter string in method SingleReturn\(bool, string\) \(x int\)`
	DoubleReturn(Empty) (x int, err error)
}
