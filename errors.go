package bind

import (
	"errors"
)

var (
	ErrReceiverUnsupportedType = errors.New("receiver was not a struct")
	ErrFieldAnonymousStruct    = errors.New("tags are not allowed with anonymous struct fields")
	ErrFieldSliceType          = errors.New("slice is not supported on field")
	ErrFieldTimeFormat         = errors.New("unable to parse time")
	ErrFieldUnsupportedType    = errors.New("unsupported type")
	ErrUnknown                 = errors.New("unknown error")
)
