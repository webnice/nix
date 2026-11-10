package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
)

const bigIntTestString = `18446744073709551616`

var (
	bigIntJSON              = []byte(`18446744073709551616`)
	bigIntStringJSON        = []byte(`"18446744073709551616"`)
	bigIntBlankJSON         = []byte(`""`)
	bigIntMaxValueValidJSON = []byte(`{"BigInt":"18446744073709551616","Valid":true}`)
	bigIntNullInvalidGob    = `30ff850301010d426967496e745772617070657201ff86000102010556616c756501ff8800010556616c696401020000000aff87050102ff8a00000003ff8600`
	bigIntOkValidGob        = `30ff850301010d426967496e745772617070657201ff86000102010556616c756501ff8800010556616c696401020000000aff87050102ff8a00000011ff86010a02010000000000000000010100`
)

func isBigIntValid(t *testing.T, b BigInt, from string) {
	const (
		errBigInt = "ошибка %q, big.Int: %q ≠ %q"
		errValid  = "является недействительным, ожидалось действительное"
	)

	if b.BigInt.String() != bigIntTestString {
		t.Errorf(errBigInt, from, b.BigInt.String(), bigIntTestString)
	}
	if !b.Valid {
		t.Error(from, errValid)
	}
}

func isBigIntNull(t *testing.T, b BigInt, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if b.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewBigInt(t *testing.T) {
	v1 := NewBigInt()
	isBigIntNull(t, v1, "NewBigInt()")
}

func TestNewBigIntValue(t *testing.T) {
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1 := NewBigIntValue(bi)
	isBigIntValid(t, v1, "NewBigIntValue()")

	v2 := NewBigIntValue(nil)
	isBigIntNull(t, v2, "NewBigIntValue(nil)")
}

func TestNewBigIntPointerValue(t *testing.T) {
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)

	v1 := NewBigIntPointerValue(&bi)
	isBigIntValid(t, v1, "NewBigIntPointerValue()")

	v2 := NewBigIntPointerValue(nil)
	isBigIntNull(t, v2, "NewBigIntPointerValue(nil)")

	v3 := NewBigIntPointerValue(new(*big.Int))
	isBigIntNull(t, v3, "NewBigIntPointerValue(*nil)")
}

func TestBigIntSetValid(t *testing.T) {
	v1 := NewBigInt()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1.SetValid(bi)
	isBigIntValid(t, v1, "SetValid()")

	v1.Set(nil)
	if !v1.Valid || v1.BigInt.Sign() != 0 {
		t.Error("SetValid(nil)", errWrong)
	}

	v2 := BigInt{}
	v2.SetValid(big.NewInt(5))
	if !v2.Valid || v2.BigInt.Int64() != 5 {
		t.Error("SetValid() для zero-value", errWrong)
	}
}

func TestBigIntReset(t *testing.T) {
	v1 := NewBigInt()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(big.NewInt(1))
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestBigIntNullIfDefault(t *testing.T) {
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1 := NewBigIntValue(bi)
	isBigIntValid(t, v1, "NewBigIntValue()")
	v1.NullIfDefault()
	isBigIntValid(t, v1, "NullIfDefault()")

	v1.Set(big.NewInt(0))
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isBigIntNull(t, v1, "NullIfDefault()")
}

func TestBigIntMustValue(t *testing.T) {
	v0 := BigInt{}
	if v0.MustValue() == nil {
		t.Error("MustValue() для zero-value", errNil)
	}
	v1 := NewBigInt()
	if v1.MustValue() == nil {
		t.Error("MustValue()", errNil)
	}
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1.SetValid(bi)
	if v1.MustValue().String() != bigIntTestString {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestBigIntPointer(t *testing.T) {
	v1 := NewBigIntPointerValue(nil)
	isBigIntNull(t, v1, "NewBigIntPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v2 := NewBigIntValue(bi)
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	pb.String() != bigIntTestString {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewBigIntPointerValue(&pb)
	isBigIntValid(t, v3, "NewBigIntPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestBigIntScan(t *testing.T) {
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)

	v1 := NewBigInt()
	errorPanic(v1.Scan(bi))
	isBigIntValid(t, v1, "Scan()")

	v2 := NewBigInt()
	errorPanic(v2.Scan(bigIntTestString))
	isBigIntValid(t, v2, "Scan()")

	v3 := NewBigInt()
	errorPanic(v3.Scan([]byte(bigIntTestString)))
	isBigIntValid(t, v3, "Scan()")

	v4 := NewBigInt()
	errorPanic(v4.Scan(int64(42)))
	if !v4.Valid || v4.BigInt.Int64() != 42 {
		t.Error("Scan()", errWrong)
	}

	v5 := NewBigInt()
	errorPanic(v5.Scan(nil))
	isBigIntNull(t, v5, "Scan()")

	v6 := NewBigInt()
	err := v6.Scan("not-a-number")
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v7 := NewBigInt()
	err = v7.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v8 := NewBigInt()
	err = v8.Scan([]byte("not-a-number"))
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestBigIntValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1 := NewBigIntValue(bi)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != bigIntTestString {
		t.Error("Value() value", errWrong)
	}

	v2 := NewBigInt()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestBigIntUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewBigInt()
	err = json.Unmarshal(bigIntJSON, &v1)
	errorPanic(err)
	isBigIntValid(t, v1, "UnmarshalJSON()")

	v2 := NewBigInt()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isBigIntNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewBigInt()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewBigInt()
	err = v4.UnmarshalJSON(bigIntMaxValueValidJSON)
	errorPanic(err)
	isBigIntValid(t, v4, "UnmarshalJSON()")

	v5 := NewBigInt()
	err = v5.UnmarshalJSON([]byte(`{"BigInt":"not-a-number","Valid":true}`))
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewBigInt()
	err = v6.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v7 := NewBigInt()
	err = v7.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v8 := NewBigInt()
	err = json.Unmarshal(bigIntStringJSON, &v8)
	errorPanic(err)
	isBigIntValid(t, v8, "UnmarshalJSON()")

	v9 := NewBigInt()
	err = json.Unmarshal(bigIntBlankJSON, &v9)
	errorPanic(err)
	isBigIntNull(t, v9, "UnmarshalJSON()")

	v10 := NewBigInt()
	err = json.Unmarshal([]byte(`"not-a-number"`), &v10)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v11 := NewBigInt()
	err = json.Unmarshal(boolFalseJSON, &v11)
	if err == nil {
		panic(errErrorNotNil)
	}
	isBigIntNull(t, v11, "UnmarshalJSON()")

	v12 := NewBigInt()
	err = json.Unmarshal([]byte(`1e100`), &v12)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Разбор в zero-value объект, как при десериализации поля структуры.
	var v13 BigInt
	err = json.Unmarshal(bigIntJSON, &v13)
	errorPanic(err)
	isBigIntValid(t, v13, "UnmarshalJSON() zero-value")

	var v14 BigInt
	err = v14.UnmarshalText([]byte(bigIntTestString))
	errorPanic(err)
	isBigIntValid(t, v14, "UnmarshalText() zero-value")
}

func TestBigIntMarshalJSON(t *testing.T) {
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1 := NewBigIntValue(bi)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(bigIntJSON), "non-empty json marshal")

	v2 := NewBigInt()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestBigIntUnmarshalText(t *testing.T) {
	var err error

	v1 := NewBigInt()
	err = v1.UnmarshalText([]byte(bigIntTestString))
	errorPanic(err)
	isBigIntValid(t, v1, "UnmarshalText()")

	v2 := NewBigInt()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.BigInt.Sign() != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewBigInt()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isBigIntNull(t, v3, "UnmarshalText()")

	v4 := NewBigInt()
	err = v4.UnmarshalText([]byte("not-a-number"))
	if err == nil {
		t.Error(errErrorNotNil)
	}
}

func TestBigIntMarshalText(t *testing.T) {
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	v1 := NewBigIntValue(bi)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, bigIntTestString, "Non-empty text marshal")

	v2 := NewBigInt()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestBigIntUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := BigInt{}
	btf, err = hex.DecodeString(bigIntNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isBigIntNull(t, v1, "UnmarshalBinary() invalid")

	v2 := BigInt{}
	btf, err = hex.DecodeString(bigIntOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isBigIntValid(t, v2, "UnmarshalBinary() ok")

	v3 := BigInt{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestBigIntMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item BigInt

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewBigInt()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isBigIntNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	bi, _ := new(big.Int).SetString(bigIntTestString, 10)
	ok := NewBigIntValue(bi)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isBigIntValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestBigIntSetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewBigInt()
	v1.SetValid(big.NewInt(0))
	isBigIntNull(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewBigInt()
	v2.Set(big.NewInt(0))
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}

	// Для zero-value объекта значение выделяется автоматически.
	var v3 BigInt
	v3.Set(big.NewInt(5))
	if !v3.Valid || v3.BigInt.Int64() != 5 {
		t.Error("Set() для zero-value", errWrong)
	}
	var v4 BigInt
	v4.SetValid(big.NewInt(5))
	if !v4.Valid || v4.BigInt.Int64() != 5 {
		t.Error("SetValid() для zero-value", errWrong)
	}

	var v5 BigInt
	v5.SetValid(nil)
	if v5.Valid || v5.BigInt.Sign() != 0 {
		t.Error("SetValid(nil)", errWrong)
	}
}
