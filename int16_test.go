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
	int16JSON              = []byte(`32767`)
	int16StringJSON        = []byte(`"32767"`)
	int16BlankJSON         = []byte(`""`)
	int16MaxValueValidJSON = []byte(`{"Int16":32767,"Valid":true}`)
	int16NullInvalidGob    = `2eff830301010c496e7431365772617070657201ff84000102010556616c7565010400010556616c6964010200000003ff8400`
	int16OkValidGob        = `2eff830301010c496e7431365772617070657201ff84000102010556616c7565010400010556616c6964010200000009ff8401fefffe010100`
)

func isInt16Valid(t *testing.T, i Int16, from string) {
	const (
		errInt16 = "ошибка %q, int16: \"%d\" ≠ \"%d\""
		errValid = "является недействительным, ожидалось действительное"
	)

	dv, _ := i.Value()
	if dv.(int64) != int64(math.MaxInt16) {
		t.Errorf(errInt16, from, dv, int64(math.MaxInt16))
	}
	if !i.Valid {
		t.Error(from, errValid)
	}
}

func isInt16Null(t *testing.T, i Int16, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewInt16(t *testing.T) {
	v1 := NewInt16()
	isInt16Null(t, v1, "NewInt16()")
}

func TestNewInt16Value(t *testing.T) {
	v1 := NewInt16Value(int16(math.MaxInt16))
	isInt16Valid(t, v1, "NewInt16Value()")
}

func TestNewInt16PointerValue(t *testing.T) {
	var bv = int16(math.MaxInt16)

	v1 := NewInt16PointerValue(&bv)
	isInt16Valid(t, v1, "NewInt16PointerValue()")

	v2 := NewInt16PointerValue(nil)
	isInt16Null(t, v2, "NewInt16PointerValue()")
}

func TestInt16SetValid(t *testing.T) {
	v1 := NewInt16()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(int16(math.MaxInt16))
	isInt16Valid(t, v1, "SetValid()")
}

func TestInt16Reset(t *testing.T) {
	v1 := NewInt16()
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

func TestInt16NullIfDefault(t *testing.T) {
	v1 := NewInt16Value(int16(math.MaxInt16))
	isInt16Valid(t, v1, "NewInt16Value()")
	v1.NullIfDefault()
	isInt16Valid(t, v1, "NullIfDefault()")

	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isInt16Null(t, v1, "NullIfDefault()")
}

func TestInt16MustValue(t *testing.T) {
	var buf any

	v1 := NewInt16()
	buf = v1.MustValue()
	if _, ok := buf.(int16); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(int16(math.MaxInt16))
	bf2 := v1.MustValue()
	if bf2 != math.MaxInt16 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestInt16Pointer(t *testing.T) {
	v1 := NewInt16PointerValue(nil)
	isInt16Null(t, v1, "NewInt16PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewInt16Value(int16(math.MaxInt16))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxInt16 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewInt16PointerValue(pb)
	isInt16Valid(t, v3, "NewInt16PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestInt16Scan(t *testing.T) {
	v1 := NewInt16()
	errorPanic(v1.Scan(int16(math.MaxInt16)))
	isInt16Valid(t, v1, "Scan()")

	v2 := NewInt16()
	errorPanic(v2.Scan("32767"))
	isInt16Valid(t, v2, "Scan()")

	v3 := NewInt16()
	errorPanic(v3.Scan(nil))
	isInt16Null(t, v3, "Scan()")

	v4 := NewInt16()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestInt16Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item int64

	v1 := NewInt16Value(int16(math.MaxInt16))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(int64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "int64")
	}
	if item != int64(math.MaxInt16) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewInt16()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestInt16UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewInt16()
	err = json.Unmarshal(int16JSON, &v1)
	errorPanic(err)
	isInt16Valid(t, v1, "UnmarshalJSON()")

	v2 := NewInt16()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isInt16Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewInt16()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewInt16()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewInt16()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewInt16()
	err = json.Unmarshal(int16StringJSON, &v6)
	errorPanic(err)
	isInt16Valid(t, v6, "UnmarshalJSON()")

	v7 := NewInt16()
	err = json.Unmarshal(int16BlankJSON, &v7)
	errorPanic(err)
	isInt16Null(t, v7, "UnmarshalJSON()")

	v8 := NewInt16()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isInt16Null(t, v8, "UnmarshalJSON()")
}

func TestInt16MarshalJSON(t *testing.T) {
	v1 := NewInt16Value(int16(math.MaxInt16))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(int16JSON), "non-empty json marshal")

	v2 := NewInt16()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestInt16UnmarshalText(t *testing.T) {
	var err error

	v1 := NewInt16()
	err = v1.UnmarshalText(int16JSON)
	errorPanic(err)
	isInt16Valid(t, v1, "UnmarshalText()")

	v2 := NewInt16()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Int16 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewInt16()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isInt16Null(t, v3, "UnmarshalText()")
}

func TestInt16MarshalText(t *testing.T) {
	v1 := NewInt16Value(int16(math.MaxInt16))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(int16JSON), "Non-empty text marshal")

	v2 := NewInt16()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestInt16UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Int16{}
	btf, err = hex.DecodeString(int16NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isInt16Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Int16{}
	btf, err = hex.DecodeString(int16OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isInt16Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Int16{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestInt16MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Int16

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewInt16()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isInt16Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewInt16Value(int16(math.MaxInt16))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isInt16Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestInt16SetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewInt16()
	v1.SetValid(0)
	isInt16Null(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewInt16()
	v2.Set(0)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
