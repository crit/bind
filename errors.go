package bind

import (
	"errors"
)

var (
	ErrReceiverUnsupportedType = errors.New("receiver was not a struct")
	ErrFieldAnonymousStruct    = errors.New("tags are not allowed with anonymous struct fields")
	ErrFieldSliceType          = errors.New("slice is not supported on field")
	ErrFieldTimeFormat         = errors.New("unable to parse time")
	ErrFieldCSVFormat          = errors.New("unable to parse csv")
	ErrFieldUnsupportedType    = errors.New("unsupported type")
	ErrFlagNotRegistered       = errors.New("flag not registered with bind.RegisterFlags")
	ErrInvalidTimeLayout       = errors.New("invalid time layout")
	ErrUnknown                 = errors.New("unknown error")
)
