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
	uint32JSON              = []byte(`4294967295`)
	uint32StringJSON        = []byte(`"4294967295"`)
	uint32BlankJSON         = []byte(`""`)
	uint32MaxValueValidJSON = []byte(`{"Uint32":4294967295,"Valid":true}`)
	uint32NullInvalidGob    = `2fff8d0301010d55696e7433325772617070657201ff8e000102010556616c7565010600010556616c6964010200000003ff8e00`
	uint32OkValidGob        = `2fff8d0301010d55696e7433325772617070657201ff8e000102010556616c7565010600010556616c696401020000000bff8e01fcffffffff010100`
)

func isUint32Valid(t *testing.T, u Uint32, from string) {
	const (
		errUint32 = "ошибка %q, uint32: \"%d\" ≠ \"%d\""
		errValid  = "является недействительным, ожидалось действительное"
	)

	dv := interface{}(u.Uint32)
	if dv.(uint32) != uint32(math.MaxUint32) {
		t.Errorf(errUint32, from, dv, uint32(math.MaxUint32))
	}
	if !u.Valid {
		t.Error(from, errValid)
	}
}

func isUint32Null(t *testing.T, i Uint32, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewUint32(t *testing.T) {
	v1 := NewUint32()
	isUint32Null(t, v1, "NewUint32()")
}

func TestNewUint32Value(t *testing.T) {
	v1 := NewUint32Value(uint32(math.MaxUint32))
	isUint32Valid(t, v1, "NewUint32Value()")
}

func TestNewUint32PointerValue(t *testing.T) {
	var bv = uint32(math.MaxUint32)

	v1 := NewUint32PointerValue(&bv)
	isUint32Valid(t, v1, "NewUint32PointerValue()")

	v2 := NewUint32PointerValue(nil)
	isUint32Null(t, v2, "NewUint32PointerValue()")
}

func TestUint32SetValid(t *testing.T) {
	v1 := NewUint32()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(uint32(math.MaxUint32))
	isUint32Valid(t, v1, "SetValid()")
}

func TestUint32Reset(t *testing.T) {
	v1 := NewUint32()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestUint32NullIfDefault(t *testing.T) {
	v1 := NewUint32Value(uint32(math.MaxUint32))
	isUint32Valid(t, v1, "NewUint32Value()")
	v1.NullIfDefault()
	isUint32Valid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isUint32Null(t, v1, "NullIfDefault()")
}

func TestUint32MustValue(t *testing.T) {
	var buf interface{}

	v1 := NewUint32()
	buf = v1.MustValue()
	if _, ok := buf.(uint32); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(uint32(math.MaxUint32))
	bf2 := v1.MustValue()
	if bf2 != math.MaxUint32 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUint32Pointer(t *testing.T) {
	v1 := NewUint32PointerValue(nil)
	isUint32Null(t, v1, "NewUint32PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewUint32Value(uint32(math.MaxUint32))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxUint32 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewUint32PointerValue(pb)
	isUint32Valid(t, v3, "NewUint32PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestUint32Scan(t *testing.T) {
	v1 := NewUint32()
	errorPanic(v1.Scan(uint32(math.MaxUint32)))
	isUint32Valid(t, v1, "Scan()")

	v2 := NewUint32()
	errorPanic(v2.Scan("4294967295"))
	isUint32Valid(t, v2, "Scan()")

	v3 := NewUint32()
	errorPanic(v3.Scan(nil))
	isUint32Null(t, v3, "Scan()")

	v4 := NewUint32()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestUint32Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	v1 := NewUint32Value(uint32(math.MaxUint32))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if !bytes.Equal(item, uint32JSON) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewUint32()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUint32UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewUint32()
	err = json.Unmarshal(uint32JSON, &v1)
	errorPanic(err)
	isUint32Valid(t, v1, "UnmarshalJSON()")

	v2 := NewUint32()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isUint32Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewUint32()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewUint32()
	err = v4.UnmarshalJSON(uint32MaxValueValidJSON)
	errorPanic(err)
	isUint32Valid(t, v4, "UnmarshalJSON()")

	v5 := NewUint32()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewUint32()
	err = json.Unmarshal(uint32StringJSON, &v6)
	errorPanic(err)
	isUint32Valid(t, v6, "UnmarshalJSON()")

	v7 := NewUint32()
	err = json.Unmarshal(uint32BlankJSON, &v7)
	errorPanic(err)
	isUint32Null(t, v7, "UnmarshalJSON()")

	v8 := NewUint32()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isUint32Null(t, v8, "UnmarshalJSON()")
}

func TestUint32MarshalJSON(t *testing.T) {
	v1 := NewUint32Value(uint32(math.MaxUint32))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(uint32JSON), "non-empty json marshal")

	v2 := NewUint32()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestUint32UnmarshalText(t *testing.T) {
	var err error

	v1 := NewUint32()
	err = v1.UnmarshalText(uint32JSON)
	errorPanic(err)
	isUint32Valid(t, v1, "UnmarshalText()")

	v2 := NewUint32()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Uint32 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewUint32()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isUint32Null(t, v3, "UnmarshalText()")
}

func TestUint32MarshalText(t *testing.T) {
	v1 := NewUint32Value(uint32(math.MaxUint32))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(uint32JSON), "Non-empty text marshal")

	v2 := NewUint32()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestUint32UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Uint32{}
	btf, err = hex.DecodeString(uint32NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isUint32Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Uint32{}
	btf, err = hex.DecodeString(uint32OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isUint32Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Uint32{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUint32MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Uint32

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewUint32()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUint32Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewUint32Value(uint32(math.MaxUint32))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUint32Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
