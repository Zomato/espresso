package utils

import (
	"context"

	"github.com/google/uuid"
)

func GenerateUniqueID(ctx context.Context) string {
	return uuid.New().String()
}
