package examples

//go:generate goderive Unique=../traits/unique

// [deriving(Unique)]
type Integers []int

// [deriving(Unique)]
type Strings []string
