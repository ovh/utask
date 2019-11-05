package sqlgenerator

import "github.com/Masterminds/squirrel"

// PGsql is a shortcut to a dollar-based statement builder
var PGsql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
