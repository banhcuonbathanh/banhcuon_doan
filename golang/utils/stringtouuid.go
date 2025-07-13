package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func StringToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func GenerateFormattedName(baseName string) string {
	now := time.Now()
	uniqueId := uuid.New().String()

	return fmt.Sprintf("%s-%d/%02d/%02d/%02d/%02d/%02d-%s",
		baseName,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		uniqueId,
	)
}
