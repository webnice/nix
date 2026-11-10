package wrapper

import "testing"

func TestExistsWrapers(t *testing.T) {
	_ = &BoolWrapper{}
	_ = &BytesWrapper{}
	_ = &Float64Wrapper{}
	_ = &Int64Wrapper{}
	_ = &StringWrapper{}
	_ = &TimeWrapper{}
	_ = &Uint64Wrapper{}
}
