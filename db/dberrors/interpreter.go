package dberrors

// ErrType describes one among possible DB-related errors
type ErrType int

// enumerate all recognized db errors
const (
	DoesNotExist ErrType = iota
	AlreadyExists
	IntegrityConstraintViolation
	InvalidInput
	Other
)

// Interpreter generates standard errors using two components:
// - an Analyzer, which is specific to a given DB (postgres, mysql, ...)
// - an ErrFactory, which generates a standard error from the ErrType output by the Analyzer
type Interpreter struct {
	Analyzer   func(error) ErrType
	ErrFactory func(error, ErrType) error
}

// Interpret converts an error from a specific db library into a different error type
func (i Interpreter) Interpret(err error) error {
	return i.ErrFactory(err, i.Analyzer(err))
}
