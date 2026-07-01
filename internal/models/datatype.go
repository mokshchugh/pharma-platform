package models

// DataType represents the type of data stored in a PLC tag.
type DataType uint8

const (
	DataTypeBool DataType = iota

	DataTypeInt16
	DataTypeUint16

	DataTypeInt32
	DataTypeUint32

	DataTypeInt64
	DataTypeUint64

	DataTypeFloat32
	DataTypeFloat64

	DataTypeString
	DataTypeBytes
)
