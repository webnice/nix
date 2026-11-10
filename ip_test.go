package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"net"
	"reflect"
	"testing"
)

const ipTestString = `192.168.0.1`

var (
	ipJSON              = []byte(`"192.168.0.1"`)
	ipBlankJSON         = []byte(`""`)
	ipMaxValueValidJSON = []byte(`{"IP":"192.168.0.1","Valid":true}`)
	ipNullInvalidGob    = `2bff830301010949505772617070657201ff84000102010556616c7565010a00010556616c6964010200000003ff8400`
	ipOkValidGob        = `2bff830301010949505772617070657201ff84000102010556616c7565010a00010556616c6964010200000017ff84011000000000000000000000ffffc0a80001010100`
)

func isIPValid(t *testing.T, i IP, from string) {
	const (
		errIP    = "ошибка %q, net.IP: %q ≠ %q"
		errValid = "является недействительным, ожидалось действительное"
	)

	if i.IP.String() != ipTestString {
		t.Errorf(errIP, from, i.IP.String(), ipTestString)
	}
	if !i.Valid {
		t.Error(from, errValid)
	}
}

func isIPNull(t *testing.T, i IP, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewIP(t *testing.T) {
	v1 := NewIP()
	isIPNull(t, v1, "NewIP()")
}

func TestNewIPValue(t *testing.T) {
	v1 := NewIPValue(net.ParseIP(ipTestString))
	isIPValid(t, v1, "NewIPValue()")
}

func TestNewIPPointerValue(t *testing.T) {
	var bv = net.ParseIP(ipTestString)

	v1 := NewIPPointerValue(&bv)
	isIPValid(t, v1, "NewIPPointerValue()")

	v2 := NewIPPointerValue(nil)
	isIPNull(t, v2, "NewIPPointerValue()")
}

func TestIPSetValid(t *testing.T) {
	v1 := NewIP()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(net.ParseIP(ipTestString))
	isIPValid(t, v1, "SetValid()")
}

func TestIPReset(t *testing.T) {
	v1 := NewIP()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(net.ParseIP("10.0.0.1"))
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestIPNullIfDefault(t *testing.T) {
	v1 := NewIPValue(net.ParseIP(ipTestString))
	isIPValid(t, v1, "NewIPValue()")
	v1.NullIfDefault()
	isIPValid(t, v1, "NullIfDefault()")

	v1.Set(nil)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isIPNull(t, v1, "NullIfDefault()")
}

func TestIPMustValue(t *testing.T) {
	v1 := NewIP()
	if v1.MustValue() != nil {
		t.Error("MustValue()", errNotNil)
	}
	v1.SetValid(net.ParseIP(ipTestString))
	if v1.MustValue().String() != ipTestString {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestIPPointer(t *testing.T) {
	v1 := NewIPPointerValue(nil)
	isIPNull(t, v1, "NewIPPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewIPValue(net.ParseIP(ipTestString))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	pb.String() != ipTestString {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewIPPointerValue(pb)
	isIPValid(t, v3, "NewIPPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestIPScan(t *testing.T) {
	v1 := NewIP()
	errorPanic(v1.Scan(net.ParseIP(ipTestString)))
	isIPValid(t, v1, "Scan()")

	v2 := NewIP()
	errorPanic(v2.Scan(ipTestString))
	isIPValid(t, v2, "Scan()")

	v3 := NewIP()
	errorPanic(v3.Scan(net.ParseIP(ipTestString).To4()))
	isIPValid(t, v3, "Scan()")

	v4 := NewIP()
	errorPanic(v4.Scan(nil))
	isIPNull(t, v4, "Scan()")

	v5 := NewIP()
	err := v5.Scan("not-an-ip-address")
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v6 := NewIP()
	err = v6.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v7 := NewIP()
	errorPanic(v7.Scan([]byte(net.ParseIP(ipTestString).To4())))
	isIPValid(t, v7, "Scan()")
}

func TestIPValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	v1 := NewIPValue(net.ParseIP(ipTestString))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != ipTestString {
		t.Error("Value() value", errWrong)
	}

	v2 := NewIP()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestIPUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewIP()
	err = json.Unmarshal(ipJSON, &v1)
	errorPanic(err)
	isIPValid(t, v1, "UnmarshalJSON()")

	v2 := NewIP()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isIPNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewIP()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewIP()
	err = v4.UnmarshalJSON(ipMaxValueValidJSON)
	errorPanic(err)
	isIPValid(t, v4, "UnmarshalJSON()")

	v5 := NewIP()
	err = v5.UnmarshalJSON([]byte(`{"IP":"192.168.0.1","Valid":false}`))
	errorPanic(err)
	isIPNull(t, v5, "UnmarshalJSON()")

	v6 := NewIP()
	err = v6.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v7 := NewIP()
	err = v7.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v8 := NewIP()
	err = json.Unmarshal(ipBlankJSON, &v8)
	errorPanic(err)
	isIPNull(t, v8, "UnmarshalJSON()")

	v9 := NewIP()
	err = json.Unmarshal([]byte(`"not-an-ip-address"`), &v9)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v10 := NewIP()
	err = json.Unmarshal(boolFalseJSON, &v10)
	if err == nil {
		panic(errErrorNotNil)
	}
	isIPNull(t, v10, "UnmarshalJSON()")
}

func TestIPMarshalJSON(t *testing.T) {
	v1 := NewIPValue(net.ParseIP(ipTestString))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(ipJSON), "non-empty json marshal")

	v2 := NewIP()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestIPUnmarshalText(t *testing.T) {
	var err error

	v1 := NewIP()
	err = v1.UnmarshalText([]byte(ipTestString))
	errorPanic(err)
	isIPValid(t, v1, "UnmarshalText()")

	v2 := NewIP()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.IP != nil || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewIP()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isIPNull(t, v3, "UnmarshalText()")

	v4 := NewIP()
	err = v4.UnmarshalText([]byte("not-an-ip-address"))
	if err == nil {
		t.Error(errErrorNotNil)
	}
}

func TestIPMarshalText(t *testing.T) {
	v1 := NewIPValue(net.ParseIP(ipTestString))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, ipTestString, "Non-empty text marshal")

	v2 := NewIP()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestIPUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := IP{}
	btf, err = hex.DecodeString(ipNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isIPNull(t, v1, "UnmarshalBinary() invalid")

	v2 := IP{}
	btf, err = hex.DecodeString(ipOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isIPValid(t, v2, "UnmarshalBinary() ok")

	v3 := IP{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestIPMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item IP

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewIP()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isIPNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewIPValue(net.ParseIP(ipTestString))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isIPValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestIPSetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewIP()
	v1.SetValid(nil)
	isIPNull(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewIP()
	v2.Set(nil)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
