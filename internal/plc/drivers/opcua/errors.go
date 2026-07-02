package opcua

import "errors"

var (
	ErrNotConnected = errors.New("opcua: client not connected")
)
