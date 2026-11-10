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
	uint8JSON              = []byte(`255`)
	uint8StringJSON        = []byte(`"255"`)
	uint8BlankJSON         = []byte(`""`)
	uint8MaxValueValidJSON = []byte(`{"Uint8":255,"Valid":true}`)
	uint8NullInvalidGob    = `2eff890301010c55696e74385772617070657201ff8a000102010556616c7565010600010556616c6964010200000003ff8a00`
	uint8OkValidGob        = `2eff890301010c55696e74385772617070657201ff8a000102010556616c7565010600010556616c6964010200000008ff8a01ffff010100`
)

func isUint8Valid(t *testing.T, u Uint8, from string) {
	const (
		errUint8 = "ошибка %q, uint8: \"%d\" ≠ \"%d\""
		errValid = "является недействительным, ожидалось действительное"
	)

	dv := any(u.Uint8)
	if dv.(uint8) != uint8(math.MaxUint8) {
		t.Errorf(errUint8, from, dv, uint8(math.MaxUint8))
	}
	if !u.Valid {
		t.Error(from, errValid)
	}
}

func isUint8Null(t *testing.T, i Uint8, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewUint8(t *testing.T) {
	v1 := NewUint8()
	isUint8Null(t, v1, "NewUint8()")
}

func TestNewUint8Value(t *testing.T) {
	v1 := NewUint8Value(uint8(math.MaxUint8))
	isUint8Valid(t, v1, "NewUint8Value()")
}

func TestNewUint8PointerValue(t *testing.T) {
	var bv = uint8(math.MaxUint8)

	v1 := NewUint8PointerValue(&bv)
	isUint8Valid(t, v1, "NewUint8PointerValue()")

	v2 := NewUint8PointerValue(nil)
	isUint8Null(t, v2, "NewUint8PointerValue()")
}

func TestUint8SetValid(t *testing.T) {
	v1 := NewUint8()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(uint8(math.MaxUint8))
	isUint8Valid(t, v1, "SetValid()")
}

func TestUint8Reset(t *testing.T) {
	v1 := NewUint8()
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

func TestUint8NullIfDefault(t *testing.T) {
	v1 := NewUint8Value(uint8(math.MaxUint8))
	isUint8Valid(t, v1, "NewUint8Value()")
	v1.NullIfDefault()
	isUint8Valid(t, v1, "NullIfDefault()")

	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isUint8Null(t, v1, "NullIfDefault()")
}

func TestUint8MustValue(t *testing.T) {
	var buf any

	v1 := NewUint8()
	buf = v1.MustValue()
	if _, ok := buf.(uint8); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(uint8(math.MaxUint8))
	bf2 := v1.MustValue()
	if bf2 != math.MaxUint8 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUint8Pointer(t *testing.T) {
	v1 := NewUint8PointerValue(nil)
	isUint8Null(t, v1, "NewUint8PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewUint8Value(uint8(math.MaxUint8))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxUint8 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewUint8PointerValue(pb)
	isUint8Valid(t, v3, "NewUint8PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestUint8Scan(t *testing.T) {
	v1 := NewUint8()
	errorPanic(v1.Scan(uint8(math.MaxUint8)))
	isUint8Valid(t, v1, "Scan()")

	v2 := NewUint8()
	errorPanic(v2.Scan("255"))
	isUint8Valid(t, v2, "Scan()")

	v3 := NewUint8()
	errorPanic(v3.Scan(nil))
	isUint8Null(t, v3, "Scan()")

	v4 := NewUint8()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestUint8Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	v1 := NewUint8Value(uint8(math.MaxUint8))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if !bytes.Equal(item, uint8JSON) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewUint8()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUint8UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewUint8()
	err = json.Unmarshal(uint8JSON, &v1)
	errorPanic(err)
	isUint8Valid(t, v1, "UnmarshalJSON()")

	v2 := NewUint8()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isUint8Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewUint8()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewUint8()
	err = v4.UnmarshalJSON(uint8MaxValueValidJSON)
	errorPanic(err)
	isUint8Valid(t, v4, "UnmarshalJSON()")

	v5 := NewUint8()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewUint8()
	err = json.Unmarshal(uint8StringJSON, &v6)
	errorPanic(err)
	isUint8Valid(t, v6, "UnmarshalJSON()")

	v7 := NewUint8()
	err = json.Unmarshal(uint8BlankJSON, &v7)
	errorPanic(err)
	isUint8Null(t, v7, "UnmarshalJSON()")

	v8 := NewUint8()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isUint8Null(t, v8, "UnmarshalJSON()")
}

func TestUint8MarshalJSON(t *testing.T) {
	v1 := NewUint8Value(uint8(math.MaxUint8))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(uint8JSON), "non-empty json marshal")

	v2 := NewUint8()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestUint8UnmarshalText(t *testing.T) {
	var err error

	v1 := NewUint8()
	err = v1.UnmarshalText(uint8JSON)
	errorPanic(err)
	isUint8Valid(t, v1, "UnmarshalText()")

	v2 := NewUint8()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Uint8 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewUint8()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isUint8Null(t, v3, "UnmarshalText()")
}

func TestUint8MarshalText(t *testing.T) {
	v1 := NewUint8Value(uint8(math.MaxUint8))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(uint8JSON), "Non-empty text marshal")

	v2 := NewUint8()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestUint8UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Uint8{}
	btf, err = hex.DecodeString(uint8NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isUint8Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Uint8{}
	btf, err = hex.DecodeString(uint8OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isUint8Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Uint8{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUint8MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Uint8

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewUint8()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUint8Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewUint8Value(uint8(math.MaxUint8))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUint8Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestUint8SetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewUint8()
	v1.SetValid(0)
	isUint8Null(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewUint8()
	v2.Set(0)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
