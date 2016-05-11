package examples

//go:generate goderive Unique=github.com/reflexionhealth/goderive/traits/unique

// [Derive(Unique)]
type Integers []int

// [Derive(Unique)]
type Strings []string
