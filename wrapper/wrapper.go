package wrapper

import (
	"encoding/gob"
	"math/big"
	"net"
	"time"
)

func init() {
	// Регистрация конкретного типа для кодера и декодера.
	gob.Register(BigIntWrapper{})
	gob.Register(BoolWrapper{})
	gob.Register(BytesWrapper{})
	gob.Register(Complex128Wrapper{})
	gob.Register(Complex64Wrapper{})
	gob.Register(DurationWrapper{})
	gob.Register(Float32Wrapper{})
	gob.Register(Float64Wrapper{})
	gob.Register(IPWrapper{})
	gob.Register(Int16Wrapper{})
	gob.Register(Int32Wrapper{})
	gob.Register(Int64Wrapper{})
	gob.Register(Int8Wrapper{})
	gob.Register(IntWrapper{})
	gob.Register(RawMessageWrapper{})
	gob.Register(StringSliceWrapper{})
	gob.Register(StringWrapper{})
	gob.Register(TimeWrapper{})
	gob.Register(Uint16Wrapper{})
	gob.Register(Uint32Wrapper{})
	gob.Register(Uint64Wrapper{})
	gob.Register(Uint8Wrapper{})
	gob.Register(UintWrapper{})
	gob.Register(UintptrWrapper{})
	gob.Register(UUIDWrapper{})
}

// BoolWrapper Обёртка для Bool.
type BoolWrapper struct {
	Value bool
	Valid bool
}

// BytesWrapper Обёртка для Bytes.
type BytesWrapper struct {
	Value []byte
	Valid bool
}

// Float64Wrapper Обёртка для Float64.
type Float64Wrapper struct {
	Value float64
	Valid bool
}

// Int64Wrapper Обёртка для Int64.
type Int64Wrapper struct {
	Value int64
	Valid bool
}

// StringSliceWrapper Обёртка для StringSlice.
type StringSliceWrapper struct {
	Value []string
	Valid bool
}

// StringWrapper Обёртка для String.
type StringWrapper struct {
	Value string
	Valid bool
}

// TimeWrapper Обёртка для Time.
type TimeWrapper struct {
	Value time.Time
	Valid bool
}

// Uint64Wrapper Обёртка для Uint64.
type Uint64Wrapper struct {
	Value uint64
	Valid bool
}

// BigIntWrapper Обёртка для BigInt.
type BigIntWrapper struct {
	Value *big.Int
	Valid bool
}

// Complex128Wrapper Обёртка для Complex128.
type Complex128Wrapper struct {
	Value complex128
	Valid bool
}

// Complex64Wrapper Обёртка для Complex64.
type Complex64Wrapper struct {
	Value complex64
	Valid bool
}

// DurationWrapper Обёртка для Duration.
type DurationWrapper struct {
	Value time.Duration
	Valid bool
}

// Float32Wrapper Обёртка для Float32.
type Float32Wrapper struct {
	Value float32
	Valid bool
}

// IPWrapper Обёртка для IP.
type IPWrapper struct {
	Value net.IP
	Valid bool
}

// Int16Wrapper Обёртка для Int16.
type Int16Wrapper struct {
	Value int16
	Valid bool
}

// Int32Wrapper Обёртка для Int32.
type Int32Wrapper struct {
	Value int32
	Valid bool
}

// Int8Wrapper Обёртка для Int8.
type Int8Wrapper struct {
	Value int8
	Valid bool
}

// IntWrapper Обёртка для Int.
type IntWrapper struct {
	Value int
	Valid bool
}

// RawMessageWrapper Обёртка для RawMessage.
type RawMessageWrapper struct {
	Value []byte
	Valid bool
}

// Uint16Wrapper Обёртка для Uint16.
type Uint16Wrapper struct {
	Value uint16
	Valid bool
}

// Uint32Wrapper Обёртка для Uint32.
type Uint32Wrapper struct {
	Value uint32
	Valid bool
}

// Uint8Wrapper Обёртка для Uint8.
type Uint8Wrapper struct {
	Value uint8
	Valid bool
}

// UintWrapper Обёртка для Uint.
type UintWrapper struct {
	Value uint
	Valid bool
}

// UintptrWrapper Обёртка для Uintptr.
type UintptrWrapper struct {
	Value uintptr
	Valid bool
}

// UUIDWrapper Обёртка для UUID.
type UUIDWrapper struct {
	Value [16]byte
	Valid bool
}
