package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

var (
	uint16JSON              = []byte(`65535`)
	uint16StringJSON        = []byte(`"65535"`)
	uint16BlankJSON         = []byte(`""`)
	uint16MaxValueValidJSON = []byte(`{"Uint16":65535,"Valid":true}`)
	uint16NullInvalidGob    = `2fff8b0301010d55696e7431365772617070657201ff8c000102010556616c7565010600010556616c6964010200000003ff8c00`
	uint16OkValidGob        = `2fff8b0301010d55696e7431365772617070657201ff8c000102010556616c7565010600010556616c6964010200000009ff8c01feffff010100`
)

func isUint16Valid(t *testing.T, u Uint16, from string) {
	const (
		errUint16 = "ошибка %q, uint16: \"%d\" ≠ \"%d\""
		errValid  = "является недействительным, ожидалось действительное"
	)

	dv := any(u.Uint16)
	if dv.(uint16) != uint16(math.MaxUint16) {
		t.Errorf(errUint16, from, dv, uint16(math.MaxUint16))
	}
	if !u.Valid {
		t.Error(from, errValid)
	}
}

func isUint16Null(t *testing.T, i Uint16, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewUint16(t *testing.T) {
	v1 := NewUint16()
	isUint16Null(t, v1, "NewUint16()")
}

func TestNewUint16Value(t *testing.T) {
	v1 := NewUint16Value(uint16(math.MaxUint16))
	isUint16Valid(t, v1, "NewUint16Value()")
}

func TestNewUint16PointerValue(t *testing.T) {
	var bv = uint16(math.MaxUint16)

	v1 := NewUint16PointerValue(&bv)
	isUint16Valid(t, v1, "NewUint16PointerValue()")

	v2 := NewUint16PointerValue(nil)
	isUint16Null(t, v2, "NewUint16PointerValue()")
}

func TestUint16SetValid(t *testing.T) {
	v1 := NewUint16()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(uint16(math.MaxUint16))
	isUint16Valid(t, v1, "SetValid()")
}

func TestUint16Reset(t *testing.T) {
	v1 := NewUint16()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestUint16NullIfDefault(t *testing.T) {
	v1 := NewUint16Value(uint16(math.MaxUint16))
	isUint16Valid(t, v1, "NewUint16Value()")
	v1.NullIfDefault()
	isUint16Valid(t, v1, "NullIfDefault()")

	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isUint16Null(t, v1, "NullIfDefault()")
}

func TestUint16MustValue(t *testing.T) {
	var buf any

	v1 := NewUint16()
	buf = v1.MustValue()
	if _, ok := buf.(uint16); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(uint16(math.MaxUint16))
	bf2 := v1.MustValue()
	if bf2 != math.MaxUint16 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUint16Pointer(t *testing.T) {
	v1 := NewUint16PointerValue(nil)
	isUint16Null(t, v1, "NewUint16PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewUint16Value(uint16(math.MaxUint16))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxUint16 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewUint16PointerValue(pb)
	isUint16Valid(t, v3, "NewUint16PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestUint16Scan(t *testing.T) {
	v1 := NewUint16()
	errorPanic(v1.Scan(uint16(math.MaxUint16)))
	isUint16Valid(t, v1, "Scan()")

	v2 := NewUint16()
	errorPanic(v2.Scan("65535"))
	isUint16Valid(t, v2, "Scan()")

	v3 := NewUint16()
	errorPanic(v3.Scan(nil))
	isUint16Null(t, v3, "Scan()")

	v4 := NewUint16()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestUint16Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	v1 := NewUint16Value(uint16(math.MaxUint16))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if !bytes.Equal(item, uint16JSON) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewUint16()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUint16UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewUint16()
	err = json.Unmarshal(uint16JSON, &v1)
	errorPanic(err)
	isUint16Valid(t, v1, "UnmarshalJSON()")

	v2 := NewUint16()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isUint16Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewUint16()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewUint16()
	err = v4.UnmarshalJSON(uint16MaxValueValidJSON)
	errorPanic(err)
	isUint16Valid(t, v4, "UnmarshalJSON()")

	v5 := NewUint16()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewUint16()
	err = json.Unmarshal(uint16StringJSON, &v6)
	errorPanic(err)
	isUint16Valid(t, v6, "UnmarshalJSON()")

	v7 := NewUint16()
	err = json.Unmarshal(uint16BlankJSON, &v7)
	errorPanic(err)
	isUint16Null(t, v7, "UnmarshalJSON()")

	v8 := NewUint16()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isUint16Null(t, v8, "UnmarshalJSON()")
}

func TestUint16MarshalJSON(t *testing.T) {
	v1 := NewUint16Value(uint16(math.MaxUint16))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(uint16JSON), "non-empty json marshal")

	v2 := NewUint16()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestUint16UnmarshalText(t *testing.T) {
	var err error

	v1 := NewUint16()
	err = v1.UnmarshalText(uint16JSON)
	errorPanic(err)
	isUint16Valid(t, v1, "UnmarshalText()")

	v2 := NewUint16()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Uint16 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewUint16()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isUint16Null(t, v3, "UnmarshalText()")
}

func TestUint16MarshalText(t *testing.T) {
	v1 := NewUint16Value(uint16(math.MaxUint16))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(uint16JSON), "Non-empty text marshal")

	v2 := NewUint16()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestUint16UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Uint16{}
	btf, err = hex.DecodeString(uint16NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isUint16Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Uint16{}
	btf, err = hex.DecodeString(uint16OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isUint16Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Uint16{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUint16MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Uint16

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewUint16()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUint16Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewUint16Value(uint16(math.MaxUint16))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUint16Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestUint16SetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewUint16()
	v1.SetValid(0)
	isUint16Null(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewUint16()
	v2.Set(0)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
