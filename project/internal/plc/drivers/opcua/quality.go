package opcua

import (
	"github.com/gopcua/opcua/ua"

	"pharma-platform/internal/models"
)

func QualityFromStatus(status ua.StatusCode) models.Quality {
	switch status {
	case ua.StatusOK:
		return models.QualityGood

	case ua.StatusUncertain:
		return models.QualityUncertain

	default:
		return models.QualityBad
	}
}
