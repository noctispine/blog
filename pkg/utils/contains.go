package utils

import (
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)


func Contains[T comparable](elems []T, v T) bool {
    for _, s := range elems {
        if v == s {
            return true
        }
    }
    return false
}

func ConstructSlug(s string) string {
    return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}

func CheckPostgreError(err error, code string) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == pgerrcode.UniqueViolation
    }

    return false
}