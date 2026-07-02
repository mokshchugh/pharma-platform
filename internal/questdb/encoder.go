package questdb

import (
	"fmt"
	"strings"

	"pharma-platform/internal/models"
)

func encode(samples []models.Sample) string {
	var b strings.Builder

	for _, sample := range samples {
		fmt.Fprintf(
			&b,
			"telemetry,plc=%s,tag=%s value=%v %d\n",
			sample.PLCID,
			sample.TagID,
			sample.Value,
			sample.Timestamp.UnixNano(),
		)
	}

	return b.String()
}
