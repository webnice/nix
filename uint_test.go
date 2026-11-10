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
	uintJSON              = []byte(`18446744073709551615`)
	uintStringJSON        = []byte(`"18446744073709551615"`)
	uintBlankJSON         = []byte(`""`)
	uintMaxValueValidJSON = []byte(`{"Uint":18446744073709551615,"Valid":true}`)
	uintNullInvalidGob    = `2dff870301010b55696e745772617070657201ff88000102010556616c7565010600010556616c6964010200000003ff8800`
	uintOkValidGob        = `2dff870301010b55696e745772617070657201ff88000102010556616c7565010600010556616c696401020000000fff8801f8ffffffffffffffff010100`
)

func isUintValid(t *testing.T, u Uint, from string) {
	const (
		errUint  = "ошибка %q, uint: \"%d\" ≠ \"%d\""
		errValid = "является недействительным, ожидалось действительное"
	)

	dv := interface{}(u.Uint)
	if dv.(uint) != uint(math.MaxUint64) {
		t.Errorf(errUint, from, dv, uint(math.MaxUint64))
	}
	if !u.Valid {
		t.Error(from, errValid)
	}
}

func isUintNull(t *testing.T, i Uint, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewUint(t *testing.T) {
	v1 := NewUint()
	isUintNull(t, v1, "NewUint()")
}

func TestNewUintValue(t *testing.T) {
	v1 := NewUintValue(uint(math.MaxUint64))
	isUintValid(t, v1, "NewUintValue()")
}

func TestNewUintPointerValue(t *testing.T) {
	var bv = uint(math.MaxUint64)

	v1 := NewUintPointerValue(&bv)
	isUintValid(t, v1, "NewUintPointerValue()")

	v2 := NewUintPointerValue(nil)
	isUintNull(t, v2, "NewUintPointerValue()")
}

func TestUintSetValid(t *testing.T) {
	v1 := NewUint()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(uint(math.MaxUint64))
	isUintValid(t, v1, "SetValid()")
}

func TestUintReset(t *testing.T) {
	v1 := NewUint()
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

func TestUintNullIfDefault(t *testing.T) {
	v1 := NewUintValue(uint(math.MaxUint64))
	isUintValid(t, v1, "NewUintValue()")
	v1.NullIfDefault()
	isUintValid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isUintNull(t, v1, "NullIfDefault()")
}

func TestUintMustValue(t *testing.T) {
	var buf interface{}

	v1 := NewUint()
	buf = v1.MustValue()
	if _, ok := buf.(uint); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(uint(math.MaxUint64))
	bf2 := v1.MustValue()
	if bf2 != math.MaxUint64 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUintPointer(t *testing.T) {
	v1 := NewUintPointerValue(nil)
	isUintNull(t, v1, "NewUintPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewUintValue(uint(math.MaxUint64))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxUint64 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewUintPointerValue(pb)
	isUintValid(t, v3, "NewUintPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestUintScan(t *testing.T) {
	v1 := NewUint()
	errorPanic(v1.Scan(uint(math.MaxUint64)))
	isUintValid(t, v1, "Scan()")

	v2 := NewUint()
	errorPanic(v2.Scan("18446744073709551615"))
	isUintValid(t, v2, "Scan()")

	v3 := NewUint()
	errorPanic(v3.Scan(nil))
	isUintNull(t, v3, "Scan()")

	v4 := NewUint()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v5 := NewUint()
	errorPanic(v5.Scan(uint64(math.MaxUint64)))
	isUintValid(t, v5, "Scan()")
}

func TestUintValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	v1 := NewUintValue(uint(math.MaxUint64))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if !bytes.Equal(item, uintJSON) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewUint()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUintUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewUint()
	err = json.Unmarshal(uintJSON, &v1)
	errorPanic(err)
	isUintValid(t, v1, "UnmarshalJSON()")

	v2 := NewUint()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isUintNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewUint()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewUint()
	err = v4.UnmarshalJSON(uintMaxValueValidJSON)
	errorPanic(err)
	isUintValid(t, v4, "UnmarshalJSON()")

	v5 := NewUint()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewUint()
	err = json.Unmarshal(uintStringJSON, &v6)
	errorPanic(err)
	isUintValid(t, v6, "UnmarshalJSON()")

	v7 := NewUint()
	err = json.Unmarshal(uintBlankJSON, &v7)
	errorPanic(err)
	isUintNull(t, v7, "UnmarshalJSON()")

	v8 := NewUint()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isUintNull(t, v8, "UnmarshalJSON()")
}

func TestUintMarshalJSON(t *testing.T) {
	v1 := NewUintValue(uint(math.MaxUint64))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(uintJSON), "non-empty json marshal")

	v2 := NewUint()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestUintUnmarshalText(t *testing.T) {
	var err error

	v1 := NewUint()
	err = v1.UnmarshalText(uintJSON)
	errorPanic(err)
	isUintValid(t, v1, "UnmarshalText()")

	v2 := NewUint()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Uint != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewUint()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isUintNull(t, v3, "UnmarshalText()")
}

func TestUintMarshalText(t *testing.T) {
	v1 := NewUintValue(uint(math.MaxUint64))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(uintJSON), "Non-empty text marshal")

	v2 := NewUint()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestUintUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Uint{}
	btf, err = hex.DecodeString(uintNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isUintNull(t, v1, "UnmarshalBinary() invalid")

	v2 := Uint{}
	btf, err = hex.DecodeString(uintOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isUintValid(t, v2, "UnmarshalBinary() ok")

	v3 := Uint{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUintMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Uint

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewUint()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUintNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewUintValue(uint(math.MaxUint64))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUintValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
