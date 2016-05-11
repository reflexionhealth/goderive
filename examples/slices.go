package examples

//go:generate goderive Unique=github.com/reflexionhealth/goderive/traits/unique

// [deriving(Unique)]
type Integers []int

// [deriving(Unique)]
type Strings []string
