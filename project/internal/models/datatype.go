package models

// DataType represents the type of data stored in a PLC tag.
type DataType uint8

const (
	DataTypeBool DataType = iota

	DataTypeInt16
	DataTypeInt32
	DataTypeInt64

	DataTypeUint16
	DataTypeUint32
	DataTypeUint64

	DataTypeFloat32
	DataTypeFloat64

	DataTypeString
	DataTypeBytes
)

func (dt DataType) String() string {
	switch dt {
	case DataTypeBool:
		return "bool"
	case DataTypeInt16:
		return "int16"
	case DataTypeInt32:
		return "int32"
	case DataTypeInt64:
		return "int64"
	case DataTypeUint16:
		return "uint16"
	case DataTypeUint32:
		return "uint32"
	case DataTypeUint64:
		return "uint64"
	case DataTypeFloat32:
		return "float32"
	case DataTypeFloat64:
		return "float64"
	case DataTypeString:
		return "string"
	case DataTypeBytes:
		return "bytes"
	default:
		return "unknown"
	}
}
