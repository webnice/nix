package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

const durationTestString = `1h30m0s`

var (
	durationJSON              = []byte(`5400000000000`)
	durationStringJSON        = []byte(`"1h30m0s"`)
	durationBlankJSON         = []byte(`""`)
	durationMaxValueValidJSON = []byte(`{"Duration":5400000000000,"Valid":true}`)
	durationNullInvalidGob    = `307f0301010f4475726174696f6e5772617070657201ff80000102010556616c7565010400010556616c6964010200000003ff8000`
	durationOkValidGob        = `307f0301010f4475726174696f6e5772617070657201ff80000102010556616c7565010400010556616c696401020000000dff8001fa09d29229e000010100`
)

func isDurationValid(t *testing.T, d Duration, from string) {
	const (
		errDuration = "ошибка %q, time.Duration: \"%d\" ≠ \"%d\""
		errValid    = "является недействительным, ожидалось действительное"
	)

	if int64(d.Duration) != int64(90*time.Minute) {
		t.Errorf(errDuration, from, int64(d.Duration), int64(90*time.Minute))
	}
	if !d.Valid {
		t.Error(from, errValid)
	}
}

func isDurationNull(t *testing.T, d Duration, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if d.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewDuration(t *testing.T) {
	v1 := NewDuration()
	isDurationNull(t, v1, "NewDuration()")
}

func TestNewDurationValue(t *testing.T) {
	v1 := NewDurationValue(90 * time.Minute)
	isDurationValid(t, v1, "NewDurationValue()")
}

func TestNewDurationPointerValue(t *testing.T) {
	var bv = 90 * time.Minute

	v1 := NewDurationPointerValue(&bv)
	isDurationValid(t, v1, "NewDurationPointerValue()")

	v2 := NewDurationPointerValue(nil)
	isDurationNull(t, v2, "NewDurationPointerValue()")
}

func TestDurationSetValid(t *testing.T) {
	v1 := NewDuration()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(90 * time.Minute)
	isDurationValid(t, v1, "SetValid()")
}

func TestDurationReset(t *testing.T) {
	v1 := NewDuration()
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

func TestDurationNullIfDefault(t *testing.T) {
	v1 := NewDurationValue(90 * time.Minute)
	isDurationValid(t, v1, "NewDurationValue()")
	v1.NullIfDefault()
	isDurationValid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isDurationNull(t, v1, "NullIfDefault()")
}

func TestDurationMustValue(t *testing.T) {
	var buf interface{}

	v1 := NewDuration()
	buf = v1.MustValue()
	if _, ok := buf.(time.Duration); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(90 * time.Minute)
	bf2 := v1.MustValue()
	if bf2 != 90*time.Minute {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestDurationPointer(t *testing.T) {
	v1 := NewDurationPointerValue(nil)
	isDurationNull(t, v1, "NewDurationPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewDurationValue(90 * time.Minute)
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != 90*time.Minute {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewDurationPointerValue(pb)
	isDurationValid(t, v3, "NewDurationPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestDurationScan(t *testing.T) {
	v1 := NewDuration()
	errorPanic(v1.Scan(90 * time.Minute))
	isDurationValid(t, v1, "Scan()")

	v2 := NewDuration()
	errorPanic(v2.Scan(int64(90 * time.Minute)))
	isDurationValid(t, v2, "Scan()")

	v3 := NewDuration()
	errorPanic(v3.Scan(durationTestString))
	isDurationValid(t, v3, "Scan()")

	v4 := NewDuration()
	errorPanic(v4.Scan(nil))
	isDurationNull(t, v4, "Scan()")

	v5 := NewDuration()
	err := v5.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestDurationValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item int64

	v1 := NewDurationValue(90 * time.Minute)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(int64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "int64")
	}
	if item != int64(90*time.Minute) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewDuration()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestDurationUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewDuration()
	err = json.Unmarshal(durationJSON, &v1)
	errorPanic(err)
	isDurationValid(t, v1, "UnmarshalJSON()")

	v2 := NewDuration()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isDurationNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewDuration()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewDuration()
	err = v4.UnmarshalJSON(durationMaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewDuration()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewDuration()
	err = json.Unmarshal(durationStringJSON, &v6)
	errorPanic(err)
	isDurationValid(t, v6, "UnmarshalJSON()")

	v7 := NewDuration()
	err = json.Unmarshal(durationBlankJSON, &v7)
	errorPanic(err)
	isDurationNull(t, v7, "UnmarshalJSON()")

	v8 := NewDuration()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isDurationNull(t, v8, "UnmarshalJSON()")
}

func TestDurationMarshalJSON(t *testing.T) {
	v1 := NewDurationValue(90 * time.Minute)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(durationJSON), "non-empty json marshal")

	v2 := NewDuration()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestDurationUnmarshalText(t *testing.T) {
	var err error

	v1 := NewDuration()
	err = v1.UnmarshalText([]byte(durationTestString))
	errorPanic(err)
	isDurationValid(t, v1, "UnmarshalText()")

	v2 := NewDuration()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Duration != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewDuration()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isDurationNull(t, v3, "UnmarshalText()")
}

func TestDurationMarshalText(t *testing.T) {
	v1 := NewDurationValue(90 * time.Minute)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, durationTestString, "Non-empty text marshal")

	v2 := NewDuration()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestDurationUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Duration{}
	btf, err = hex.DecodeString(durationNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isDurationNull(t, v1, "UnmarshalBinary() invalid")

	v2 := Duration{}
	btf, err = hex.DecodeString(durationOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isDurationValid(t, v2, "UnmarshalBinary() ok")

	v3 := Duration{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestDurationMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Duration

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewDuration()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isDurationNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewDurationValue(90 * time.Minute)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isDurationValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
