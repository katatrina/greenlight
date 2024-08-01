package util

import "github.com/katatrina/greenlight/internal/db"

func GetNullableString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func GetNullableInt32(i *int32) int32 {
	if i == nil {
		return 0
	}

	return *i
}

func GetNullableRuntime(i *db.Runtime) db.Runtime {
	if i == nil {
		return 0
	}

	return *i
}
