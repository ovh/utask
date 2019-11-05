package pgjuju

import (
	"github.com/juju/errors"
	"github.com/ovh/utask/db/dberrors"
	"github.com/ovh/utask/db/dberrors/pganalyzer"
)

var interpreter = dberrors.Interpreter{
	Analyzer: pganalyzer.Analyzer,
	ErrFactory: func(err error, errType dberrors.ErrType) error {
		switch errType {
		case dberrors.DoesNotExist:
			return errors.NewNotFound(err, "")
		case dberrors.InvalidInput, dberrors.IntegrityConstraintViolation:
			return errors.NewNotValid(err, "")
		case dberrors.AlreadyExists:
			return errors.NewAlreadyExists(err, "")
		case dberrors.Other:
			return err
		}
		return err
	},
}

// Interpret converts a postgresql error into a juju error,
// convertible by the API server into a status code
func Interpret(err error) error {
	return interpreter.Interpret(err)
}
