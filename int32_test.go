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
	int32JSON              = []byte(`2147483647`)
	int32StringJSON        = []byte(`"2147483647"`)
	int32BlankJSON         = []byte(`""`)
	int32MaxValueValidJSON = []byte(`{"Int32":2147483647,"Valid":true}`)
	int32NullInvalidGob    = `2eff850301010c496e7433325772617070657201ff86000102010556616c7565010400010556616c6964010200000003ff8600`
	int32OkValidGob        = `2eff850301010c496e7433325772617070657201ff86000102010556616c7565010400010556616c696401020000000bff8601fcfffffffe010100`
)

func isInt32Valid(t *testing.T, i Int32, from string) {
	const (
		errInt32 = "ошибка %q, int32: \"%d\" ≠ \"%d\""
		errValid = "является недействительным, ожидалось действительное"
	)

	dv, _ := i.Value()
	if dv.(int64) != int64(math.MaxInt32) {
		t.Errorf(errInt32, from, dv, int64(math.MaxInt32))
	}
	if !i.Valid {
		t.Error(from, errValid)
	}
}

func isInt32Null(t *testing.T, i Int32, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewInt32(t *testing.T) {
	v1 := NewInt32()
	isInt32Null(t, v1, "NewInt32()")
}

func TestNewInt32Value(t *testing.T) {
	v1 := NewInt32Value(int32(math.MaxInt32))
	isInt32Valid(t, v1, "NewInt32Value()")
}

func TestNewInt32PointerValue(t *testing.T) {
	var bv = int32(math.MaxInt32)

	v1 := NewInt32PointerValue(&bv)
	isInt32Valid(t, v1, "NewInt32PointerValue()")

	v2 := NewInt32PointerValue(nil)
	isInt32Null(t, v2, "NewInt32PointerValue()")
}

func TestInt32SetValid(t *testing.T) {
	v1 := NewInt32()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(int32(math.MaxInt32))
	isInt32Valid(t, v1, "SetValid()")
}

func TestInt32Reset(t *testing.T) {
	v1 := NewInt32()
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

func TestInt32NullIfDefault(t *testing.T) {
	v1 := NewInt32Value(int32(math.MaxInt32))
	isInt32Valid(t, v1, "NewInt32Value()")
	v1.NullIfDefault()
	isInt32Valid(t, v1, "NullIfDefault()")

	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isInt32Null(t, v1, "NullIfDefault()")
}

func TestInt32MustValue(t *testing.T) {
	var buf any

	v1 := NewInt32()
	buf = v1.MustValue()
	if _, ok := buf.(int32); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(int32(math.MaxInt32))
	bf2 := v1.MustValue()
	if bf2 != math.MaxInt32 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestInt32Pointer(t *testing.T) {
	v1 := NewInt32PointerValue(nil)
	isInt32Null(t, v1, "NewInt32PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewInt32Value(int32(math.MaxInt32))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxInt32 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewInt32PointerValue(pb)
	isInt32Valid(t, v3, "NewInt32PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestInt32Scan(t *testing.T) {
	v1 := NewInt32()
	errorPanic(v1.Scan(int32(math.MaxInt32)))
	isInt32Valid(t, v1, "Scan()")

	v2 := NewInt32()
	errorPanic(v2.Scan("2147483647"))
	isInt32Valid(t, v2, "Scan()")

	v3 := NewInt32()
	errorPanic(v3.Scan(nil))
	isInt32Null(t, v3, "Scan()")

	v4 := NewInt32()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestInt32Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item int64

	v1 := NewInt32Value(int32(math.MaxInt32))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(int64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "int64")
	}
	if item != int64(math.MaxInt32) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewInt32()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestInt32UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewInt32()
	err = json.Unmarshal(int32JSON, &v1)
	errorPanic(err)
	isInt32Valid(t, v1, "UnmarshalJSON()")

	v2 := NewInt32()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isInt32Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewInt32()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewInt32()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewInt32()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewInt32()
	err = json.Unmarshal(int32StringJSON, &v6)
	errorPanic(err)
	isInt32Valid(t, v6, "UnmarshalJSON()")

	v7 := NewInt32()
	err = json.Unmarshal(int32BlankJSON, &v7)
	errorPanic(err)
	isInt32Null(t, v7, "UnmarshalJSON()")

	v8 := NewInt32()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isInt32Null(t, v8, "UnmarshalJSON()")
}

func TestInt32MarshalJSON(t *testing.T) {
	v1 := NewInt32Value(int32(math.MaxInt32))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(int32JSON), "non-empty json marshal")

	v2 := NewInt32()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestInt32UnmarshalText(t *testing.T) {
	var err error

	v1 := NewInt32()
	err = v1.UnmarshalText(int32JSON)
	errorPanic(err)
	isInt32Valid(t, v1, "UnmarshalText()")

	v2 := NewInt32()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Int32 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewInt32()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isInt32Null(t, v3, "UnmarshalText()")
}

func TestInt32MarshalText(t *testing.T) {
	v1 := NewInt32Value(int32(math.MaxInt32))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(int32JSON), "Non-empty text marshal")

	v2 := NewInt32()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestInt32UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Int32{}
	btf, err = hex.DecodeString(int32NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isInt32Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Int32{}
	btf, err = hex.DecodeString(int32OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isInt32Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Int32{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestInt32MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Int32

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewInt32()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isInt32Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewInt32Value(int32(math.MaxInt32))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isInt32Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestInt32SetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewInt32()
	v1.SetValid(0)
	isInt32Null(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewInt32()
	v2.Set(0)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
