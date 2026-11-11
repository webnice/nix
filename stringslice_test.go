package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// stringSliceTestBody Тестовое значение типа []string, строка формата set.
var stringSliceTestBody = []string{"один", "два", "три", "четыре", "пять"}

func isStringSliceValid(t *testing.T, ss StringSlice, from string) {
	const errString = "ошибка %q, значение: \"%s\" ≠ \"%s\""

	dv, _ := ss.Value()
	if dv.(string) != strings.Join(stringSliceTestBody, ",") {
		t.Errorf(errString, from, dv, strings.Join(stringSliceTestBody, ","))
	}
	if !ss.Valid {
		t.Error(from, errValid)
	}
}

func isStringSliceNull(t *testing.T, ss StringSlice, from string) {
	if ss.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewStringSlice(t *testing.T) {
	v1 := NewStringSlice()
	isStringSliceNull(t, v1, "NewStringSlice()")
}

func TestNewStringSliceValue(t *testing.T) {
	v1 := NewStringSliceValue(stringSliceTestBody)
	isStringSliceValid(t, v1, "NewStringSliceValue()")
}

func TestNewStringSlicePointerValue(t *testing.T) {
	var bv = stringSliceTestBody

	v1 := NewStringSlicePointerValue(&bv)
	isStringSliceValid(t, v1, "NewStringSlicePointerValue()")

	v2 := NewStringSlicePointerValue(nil)
	isStringSliceNull(t, v2, "NewStringSlicePointerValue()")
}

func TestStringSliceSetValid(t *testing.T) {
	v1 := NewStringSlice()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(stringSliceTestBody)
	isStringSliceValid(t, v1, "SetValid()")
}

func TestStringSliceReset(t *testing.T) {
	v1 := NewStringSlice()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.Set([]string{})
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
}

func TestStringSliceNullIfDefault(t *testing.T) {
	v1 := NewStringSliceValue(stringSliceTestBody)
	isStringSliceValid(t, v1, "NewStringSliceValue()")
	v1.NullIfDefault()
	isStringSliceValid(t, v1, "NullIfDefault()")

	v1.Set([]string{})
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.NullIfDefault()
	isStringSliceNull(t, v1, "NullIfDefault()")
}

func TestStringSliceMustValue(t *testing.T) {
	var buf any

	v1 := NewStringSlice()
	buf = v1.MustValue()
	if _, ok := buf.([]string); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(stringSliceTestBody)
	bf2 := v1.MustValue()
	if !reflect.DeepEqual(bf2, stringSliceTestBody) {
		t.Error("MustValue()", errWrong)
	}
}

func TestStringSlicePointer(t *testing.T) {
	v1 := NewStringSlicePointerValue(nil)
	isStringSliceNull(t, v1, "NewStringSlicePointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewStringSliceValue(stringSliceTestBody)
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if !reflect.DeepEqual(*pb, stringSliceTestBody) {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewStringSlicePointerValue(pb)
	isStringSliceValid(t, v3, "NewStringSlicePointerValue(pb)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestStringSliceScan(t *testing.T) {
	// Сканирование строки формата set базы данных.
	v1 := NewStringSlice()
	errorPanic(v1.Scan(strings.Join(stringSliceTestBody, ",")))
	isStringSliceValid(t, v1, "Scan(string)")

	// Драйверы БД передают строковые значения также в виде []byte.
	v2 := NewStringSlice()
	errorPanic(v2.Scan([]byte(strings.Join(stringSliceTestBody, ","))))
	isStringSliceValid(t, v2, "Scan([]byte)")

	// Пустая строка — пустой набор set, значение действительное.
	v3 := NewStringSlice()
	errorPanic(v3.Scan(""))
	if !v3.Valid {
		t.Error("Scan(\"\")", errValidFalse)
	}
	if len(v3.StringSlice) != 0 {
		t.Error("Scan(\"\")", errWrong)
	}

	// NULL — нулевой объект.
	v4 := NewStringSlice()
	errorPanic(v4.Scan(nil))
	isStringSliceNull(t, v4, "Scan(nil)")

	// Неподдерживаемый тип — ошибка.
	v5 := NewStringSlice()
	if err := v5.Scan(42); err == nil {
		t.Error("Scan(42)", errErrorNotNil)
	}
}

func TestStringSliceValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	v1 := NewStringSliceValue(stringSliceTestBody)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != strings.Join(stringSliceTestBody, ",") {
		t.Error("Value() value", errWrong)
	}

	v2 := NewStringSlice()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestStringSliceUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewStringSlice()
	err = json.Unmarshal([]byte(`["один","два","три","четыре","пять"]`), &v1)
	errorPanic(err)
	isStringSliceValid(t, v1, "UnmarshalJSON()")

	v2 := NewStringSlice()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isStringSliceNull(t, v2, "UnmarshalJSON(null)")

	// Объект {"StringSlice":…,"Valid":…}.
	v3 := NewStringSlice()
	err = json.Unmarshal([]byte(`{"StringSlice":["a","b"],"Valid":true}`), &v3)
	errorPanic(err)
	if !v3.Valid || !reflect.DeepEqual(v3.StringSlice, []string{"a", "b"}) {
		t.Error("UnmarshalJSON(object)", errValueValid)
	}

	// Объект с Valid=false.
	v4 := NewStringSlice()
	err = json.Unmarshal([]byte(`{"StringSlice":["a"],"Valid":false}`), &v4)
	errorPanic(err)
	isStringSliceNull(t, v4, "UnmarshalJSON(object invalid)")

	// Объект с ключом StringSlice не того типа.
	v41 := NewStringSlice()
	err = json.Unmarshal([]byte(`{"StringSlice":"a,b","Valid":true}`), &v41)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Объект с ключом Valid не того типа.
	v42 := NewStringSlice()
	err = json.Unmarshal([]byte(`{"StringSlice":["a"],"Valid":"true"}`), &v42)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Пустой массив — действительное пустое значение.
	v5 := NewStringSlice()
	err = json.Unmarshal([]byte(`[]`), &v5)
	errorPanic(err)
	if !v5.Valid || len(v5.StringSlice) != 0 {
		t.Error("UnmarshalJSON([])", errValueValid)
	}

	// Синтаксическая ошибка JSON.
	v6 := NewStringSlice()
	err = v6.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	// Элемент массива не является строкой.
	v7 := NewStringSlice()
	err = v7.UnmarshalJSON([]byte(`["a",1]`))
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Значение не является массивом или объектом.
	v8 := NewStringSlice()
	err = json.Unmarshal(int64JSON, &v8)
	if err == nil {
		t.Error(errErrorNotNil)
	}
	isStringSliceNull(t, v8, "UnmarshalJSON(int64)")
}

func TestStringSliceMarshalJSON(t *testing.T) {
	v1 := NewStringSliceValue(stringSliceTestBody)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, `["один","два","три","четыре","пять"]`, "non-empty json marshal")

	v2 := NewStringSlice()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

func TestStringSliceUnmarshalText(t *testing.T) {
	var err error

	v1 := NewStringSlice()
	err = v1.UnmarshalText([]byte(strings.Join(stringSliceTestBody, ",")))
	errorPanic(err)
	isStringSliceValid(t, v1, "UnmarshalText()")

	v2 := NewStringSlice()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if len(v2.StringSlice) != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewStringSlice()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isStringSliceNull(t, v3, "UnmarshalText()")
}

func TestStringSliceMarshalText(t *testing.T) {
	v1 := NewStringSliceValue(stringSliceTestBody)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, strings.Join(stringSliceTestBody, ","), "Non-empty text marshal")

	v2 := NewStringSlice()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestStringSliceUnmarshalBinary(t *testing.T) {
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item StringSlice

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	null := NewStringSlice()
	errorPanic(enc.Encode(&null))
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	errorPanic(gob.NewDecoder(buf).Decode(&item))
	isStringSliceNull(t, item, "UnmarshalBinary() invalid")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewStringSliceValue(stringSliceTestBody)
	errorPanic(enc.Encode(&ok))
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	errorPanic(gob.NewDecoder(buf).Decode(&item))
	isStringSliceValid(t, item, "UnmarshalBinary() ok")
}

//goland:noinspection DuplicatedCode
func TestStringSliceMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item StringSlice

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewStringSlice()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isStringSliceNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewStringSliceValue(stringSliceTestBody)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isStringSliceValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestStringSliceSetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewStringSlice()
	v1.SetValid([]string{})
	isStringSliceNull(t, v1, "SetValid(default)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewStringSlice()
	v2.Set([]string{})
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}

// stringSliceTestComma Тестовое значение с запятыми и обратными косыми
// чертами в элементах.
var stringSliceTestComma = []string{`раз, два`, `три\четыре`, `пять\, шесть`}

// stringSliceTestCommaSet Строковое представление значения stringSliceTestComma
// в формате set базы данных.
const stringSliceTestCommaSet = `раз\, два,три\\четыре,пять\\\, шесть`

func TestStringSliceEscape(t *testing.T) {
	var data []byte
	var err error

	// Экранирование при записи: запятая и обратная косая черта кодируются
	// обратной косой чертой.
	v1 := NewStringSliceValue(stringSliceTestComma)
	dv, err := v1.Value()
	errorPanic(err)
	jsonEquals(t, []byte(dv.(string)), stringSliceTestCommaSet, "Value() escape")

	data, err = v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, stringSliceTestCommaSet, "MarshalText() escape")

	// Восстановление экранирования при чтении из строки и []byte.
	v2 := NewStringSlice()
	errorPanic(v2.Scan(stringSliceTestCommaSet))
	if !reflect.DeepEqual(v2.StringSlice, stringSliceTestComma) {
		t.Error("Scan() escape", errWrong)
	}

	v3 := NewStringSlice()
	errorPanic(v3.Scan([]byte(stringSliceTestCommaSet)))
	if !reflect.DeepEqual(v3.StringSlice, stringSliceTestComma) {
		t.Error("Scan([]byte) escape", errWrong)
	}

	// Восстановление экранирования при разборе текстового представления.
	v4 := NewStringSlice()
	errorPanic(v4.UnmarshalText([]byte(stringSliceTestCommaSet)))
	if !reflect.DeepEqual(v4.StringSlice, stringSliceTestComma) {
		t.Error("UnmarshalText() escape", errWrong)
	}

	// Текстовый цикл: MarshalText -> UnmarshalText.
	v5 := NewStringSliceValue(stringSliceTestComma)
	data, err = v5.MarshalText()
	errorPanic(err)
	v6 := NewStringSlice()
	errorPanic(v6.UnmarshalText(data))
	if !reflect.DeepEqual(v6.StringSlice, stringSliceTestComma) {
		t.Error("MarshalText() -> UnmarshalText() escape", errWrong)
	}

	// Незнакомые последовательности экранирования и завершающая косая
	// черта сохраняются как есть.
	v7 := NewStringSlice()
	errorPanic(v7.Scan(`a\bc,a\`))
	if !reflect.DeepEqual(v7.StringSlice, []string{`a\bc`, `a\`}) {
		t.Error("Scan() незнакомое экранирование", errWrong)
	}

	// JSON не экранирует: элементы передаются в массиве как есть.
	v8 := NewStringSlice()
	errorPanic(json.Unmarshal([]byte(`["раз, два","три"]`), &v8))
	if !reflect.DeepEqual(v8.StringSlice, []string{"раз, два", "три"}) {
		t.Error("UnmarshalJSON() запятая", errWrong)
	}
}
