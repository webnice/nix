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
	float32JSON              = []byte(`340282346638528859811704183484516925440.000000`)
	float32StringJSON        = []byte(`"340282346638528859811704183484516925440.000000"`)
	float32BlankJSON         = []byte(`""`)
	float32MaxValueValidJSON = []byte(`{"Float32":340282346638528859811704183484516925440.000000,"Valid":true}`)
	float32NullInvalidGob    = `30ff910301010e466c6f617433325772617070657201ff92000102010556616c7565010800010556616c6964010200000003ff9200`
	float32OkValidGob        = `30ff910301010e466c6f617433325772617070657201ff92000102010556616c7565010800010556616c696401020000000cff9201fbe0ffffef47010100`
)

func isFloat32Valid(t *testing.T, f Float32, from string) {
	const (
		errFloat32 = "ошибка %q, float32: \"%f\" ≠ \"%f\""
		errValid   = "является недействительным, ожидалось действительное"
	)

	dv, _ := f.Value()
	if dv.(float64) != float64(math.MaxFloat32) {
		t.Errorf(errFloat32, from, dv, float64(math.MaxFloat32))
	}
	if !f.Valid {
		t.Error(from, errValid)
	}
}

func isFloat32Null(t *testing.T, i Float32, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewFloat32(t *testing.T) {
	v1 := NewFloat32()
	isFloat32Null(t, v1, "NewFloat32()")
}

func TestNewFloat32Value(t *testing.T) {
	v1 := NewFloat32Value(float32(math.MaxFloat32))
	isFloat32Valid(t, v1, "NewFloat32Value()")
}

func TestNewFloat32PointerValue(t *testing.T) {
	var bv = float32(math.MaxFloat32)

	v1 := NewFloat32PointerValue(&bv)
	isFloat32Valid(t, v1, "NewFloat32PointerValue()")

	v2 := NewFloat32PointerValue(nil)
	isFloat32Null(t, v2, "NewFloat32PointerValue()")
}

func TestFloat32SetValid(t *testing.T) {
	v1 := NewFloat32()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(float32(math.MaxFloat32))
	isFloat32Valid(t, v1, "SetValid()")
}

func TestFloat32Reset(t *testing.T) {
	v1 := NewFloat32()
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

func TestFloat32NullIfDefault(t *testing.T) {
	v1 := NewFloat32Value(float32(math.MaxFloat32))
	isFloat32Valid(t, v1, "NewFloat32Value()")
	v1.NullIfDefault()
	isFloat32Valid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isFloat32Null(t, v1, "NullIfDefault()")
}

func TestFloat32MustValue(t *testing.T) {
	var buf interface{}

	v1 := NewFloat32()
	buf = v1.MustValue()
	if _, ok := buf.(float32); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(float32(math.MaxFloat32))
	bf2 := v1.MustValue()
	if bf2 != math.MaxFloat32 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat32Pointer(t *testing.T) {
	v1 := NewFloat32PointerValue(nil)
	isFloat32Null(t, v1, "NewFloat32PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewFloat32Value(float32(math.MaxFloat32))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxFloat32 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewFloat32PointerValue(pb)
	isFloat32Valid(t, v3, "NewFloat32PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestFloat32Scan(t *testing.T) {
	v1 := NewFloat32()
	errorPanic(v1.Scan(float32(math.MaxFloat32)))
	isFloat32Valid(t, v1, "Scan()")

	v2 := NewFloat32()
	errorPanic(v2.Scan("340282346638528859811704183484516925440.000000"))
	isFloat32Valid(t, v2, "Scan()")

	v3 := NewFloat32()
	errorPanic(v3.Scan(nil))
	isFloat32Null(t, v3, "Scan()")

	v4 := NewFloat32()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v5 := NewFloat32()
	errorPanic(v5.Scan(float64(math.MaxFloat32)))
	isFloat32Valid(t, v5, "Scan()")
}

func TestFloat32Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item float64

	v1 := NewFloat32Value(float32(math.MaxFloat32))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(float64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "float64")
	}
	if item != float64(math.MaxFloat32) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewFloat32()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat32UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewFloat32()
	err = json.Unmarshal(float32JSON, &v1)
	errorPanic(err)
	isFloat32Valid(t, v1, "UnmarshalJSON()")

	v2 := NewFloat32()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isFloat32Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewFloat32()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewFloat32()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewFloat32()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewFloat32()
	err = json.Unmarshal(float32StringJSON, &v6)
	errorPanic(err)
	isFloat32Valid(t, v6, "UnmarshalJSON()")

	v7 := NewFloat32()
	err = json.Unmarshal(float32BlankJSON, &v7)
	errorPanic(err)
	isFloat32Null(t, v7, "UnmarshalJSON()")

	v8 := NewFloat32()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isFloat32Null(t, v8, "UnmarshalJSON()")
}

func TestFloat32MarshalJSON(t *testing.T) {
	v1 := NewFloat32Value(float32(math.MaxFloat32))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(float32JSON), "non-empty json marshal")

	v2 := NewFloat32()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")

	v3 := NewFloat32Value(float32(math.NaN()))
	data, err = v3.MarshalJSON()
	if err == nil || len(data) > 0 {
		panic(errErrorNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat32UnmarshalText(t *testing.T) {
	var err error

	v1 := NewFloat32()
	err = v1.UnmarshalText(float32JSON)
	errorPanic(err)
	isFloat32Valid(t, v1, "UnmarshalText()")

	v2 := NewFloat32()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Float32 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewFloat32()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isFloat32Null(t, v3, "UnmarshalText()")
}

func TestFloat32MarshalText(t *testing.T) {
	v1 := NewFloat32Value(float32(math.MaxFloat32))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(float32JSON), "Non-empty text marshal")

	v2 := NewFloat32()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestFloat32UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Float32{}
	btf, err = hex.DecodeString(float32NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isFloat32Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Float32{}
	btf, err = hex.DecodeString(float32OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isFloat32Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Float32{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat32MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Float32

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewFloat32()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isFloat32Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewFloat32Value(float32(math.MaxFloat32))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isFloat32Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
