package pganalyzer

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/ovh/utask/db/dberrors"
)

// Analyzer converts a postgresql-specific error
// into a generic dberrors.ErrType
var Analyzer = func(err error) dberrors.ErrType {
	if err == sql.ErrNoRows {
		return dberrors.DoesNotExist
	}
	pgErr, ok := err.(*pq.Error)
	if !ok {
		return dberrors.Other
	}
	// Unique constraint violation errors
	if pgErr.Code.Name() == "unique_violation" {
		return dberrors.AlreadyExists
	}
	// Constraints violation errors
	if pgErr.Code.Class().Name() == "integrity_constraint_violation" {
		return dberrors.IntegrityConstraintViolation
	}
	// Data error
	if pgErr.Code.Class().Name() == "data_exception" {
		return dberrors.InvalidInput
	}
	return dberrors.Other
}
