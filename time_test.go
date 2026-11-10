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

func isTimeValid(t *testing.T, tm Time, from string) {
	const errTime = "ошибка %q, time.Time: \"%s\" ≠ \"%s\""

	dv, _ := tm.Value()
	if !dv.(time.Time).Equal(timeOkValidValue) {
		t.Errorf(errTime, from, dv, timeOkValidValue)
	}
	if !tm.Valid {
		t.Error(from, errValid)
	}
}

func isTimeNull(t *testing.T, tm Time, from string) {
	if tm.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewTime(t *testing.T) {
	v1 := NewTime()
	isTimeNull(t, v1, "NewTime()")
}

func TestNewTimeValue(t *testing.T) {
	v1 := NewTimeValue(timeOkValidValue)
	isTimeValid(t, v1, "NewTimeValue()")
}

func TestNewTimePointerValue(t *testing.T) {
	var bv = timeOkValidValue

	v1 := NewTimePointerValue(&bv)
	isTimeValid(t, v1, "NewTimePointerValue()")

	v2 := NewTimePointerValue(nil)
	isTimeNull(t, v2, "NewTimePointerValue()")
}

func TestTimeSetValid(t *testing.T) {
	v1 := NewTime()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(timeOkValidValue)
	isTimeValid(t, v1, "SetValid()")
}

func TestTimeReset(t *testing.T) {
	v1 := NewTime()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(time.Time{})
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
}

func TestTimeNullIfDefault(t *testing.T) {
	v1 := NewTimeValue(timeOkValidValue)
	isTimeValid(t, v1, "NewTimeValue()")
	v1.NullIfDefault()
	isTimeValid(t, v1, "NullIfDefault()")

	v1.SetValid(time.Time{})
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.NullIfDefault()
	isTimeNull(t, v1, "NullIfDefault()")
}

func TestTimeMustValue(t *testing.T) {
	var buf interface{}

	v1 := NewTime()
	buf = v1.MustValue()
	if _, ok := buf.(time.Time); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(timeOkValidValue)
	bf2 := v1.MustValue()
	if !bf2.Equal(timeOkValidValue) {
		t.Error("MustValue()", errWrong)
	}
}

func TestTimePointer(t *testing.T) {
	v1 := NewTimePointerValue(nil)
	isTimeNull(t, v1, "NewTimePointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	var pb *time.Time
	v2 := NewTimeValue(timeOkValidValue)
	pb = v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	!pb.Equal(timeOkValidValue) {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewTimePointerValue(pb)
	isTimeValid(t, v3, "NewTimePointerValue()")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestTimeScan(t *testing.T) {
	v1 := NewTime()
	errorPanic(v1.Scan(timeOkValidValue))
	isTimeValid(t, v1, "Scan()")

	v2 := NewTime()
	errorPanic(v2.Scan(timeStringValue))
	isTimeValid(t, v2, "Scan()")

	v3 := NewTime()
	errorPanic(v3.Scan(nil))
	isTimeNull(t, v3, "Scan()")

	v4 := NewTime()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestTimeValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item time.Time

	v1 := NewTimeValue(timeOkValidValue)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(time.Time); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "time.Time")
	}
	if !item.Equal(timeOkValidValue) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewTime()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestTimeUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewTime()
	err = json.Unmarshal(timeStringValueJSON, &v1)
	errorPanic(err)
	isTimeValid(t, v1, "UnmarshalJSON()")

	v2 := NewTime()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isTimeNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewTime()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewTime()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewTime()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewTime()
	err = json.Unmarshal(boolFalseJSON, &v6)
	if err == nil {
		panic(errErrorNotNil)
	}
	isTimeNull(t, v6, "UnmarshalJSON()")

	v7 := NewTime()
	err = json.Unmarshal(timeTestValidJSON, &v7)
	errorPanic(err)
	isTimeValid(t, v7, "UnmarshalJSON()")
}

func TestTimeMarshalJSON(t *testing.T) {
	v1 := NewTimeValue(timeOkValidValue)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(timeStringValueJSON), "non-empty json marshal")

	v2 := NewTime()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

func TestTimeUnmarshalText(t *testing.T) {
	var err error

	v1 := NewTime()
	err = v1.UnmarshalText([]byte(timeStringValue))
	errorPanic(err)
	isTimeValid(t, v1, "UnmarshalText()")

	v2 := NewTime()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if !v2.Time.IsZero() || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewTime()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isTimeNull(t, v3, "UnmarshalText()")
}

func TestTimeMarshalText(t *testing.T) {
	v1 := NewTimeValue(timeOkValidValue)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, timeStringValue, "Non-empty text marshal")

	v2 := NewTime()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")

	v3 := NewTimeValue(time.Time{})
	data, err = v3.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestTimeUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Time{}
	btf, err = hex.DecodeString(timeNullInvalidGob)
	errorPanic(err)
	errorPanic(gob.NewDecoder(bytes.NewBuffer(btf)).Decode(&v1))
	isTimeNull(t, v1, "UnmarshalBinary() invalid")

	v2 := Time{}
	btf, err = hex.DecodeString(timeOkValidGob)
	errorPanic(err)
	errorPanic(gob.NewDecoder(bytes.NewBuffer(btf)).Decode(&v2))
	isTimeValid(t, v2, "UnmarshalBinary() ok")
}

//goland:noinspection DuplicatedCode
func TestTimeMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Time

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewTime()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isTimeNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewTimeValue(timeOkValidValue)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isTimeValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
