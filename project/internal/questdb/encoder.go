package questdb

import (
	"fmt"
	"strings"

	"pharma-platform/internal/models"
)

func encode(
	table string,
	samples []models.Sample,
) string {
	var b strings.Builder

	for _, sample := range samples {
		fmt.Fprintf(
			&b,
			"%s,machine_id=%s,machine_name=%s,tag_name=%s value=%s,quality=%di %d\n",
			table,
			escapeSymbol(sample.MachineID),
			escapeSymbol(sample.MachineName),
			escapeSymbol(sample.TagName),
			encodeValue(sample.Value),
			sample.Quality,
			sample.Timestamp.UnixNano(),
		)
	}

	return b.String()
}

func encodeValue(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"

	case int:
		return fmt.Sprintf("%d.0", v)

	case int8:
		return fmt.Sprintf("%d.0", v)

	case int16:
		return fmt.Sprintf("%d.0", v)

	case int32:
		return fmt.Sprintf("%d.0", v)

	case int64:
		return fmt.Sprintf("%d.0", v)

	case uint:
		return fmt.Sprintf("%d.0", v)

	case uint8:
		return fmt.Sprintf("%d.0", v)

	case uint16:
		return fmt.Sprintf("%d.0", v)

	case uint32:
		return fmt.Sprintf("%d.0", v)

	case uint64:
		return fmt.Sprintf("%d.0", v)

	case float32:
		return fmt.Sprintf("%f", v)

	case float64:
		return fmt.Sprintf("%f", v)

	case string:
		return fmt.Sprintf("\"%s\"", v)

	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

var symbolEscaper = strings.NewReplacer(
	"\\", "\\\\",
	" ", "\\ ",
	",", "\\,",
	"=", "\\=",
)

func escapeSymbol(s string) string {
	return symbolEscaper.Replace(s)
}
