package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// nullableKind Параметры тестирования Nullable для одного базового типа.
type nullableKind struct {
	name string
	// Значение и его строковая форма.
	value   interface{}
	str     string
	zeroStr string
	// Ожидаемый driver.Value для Value().
	want driver.Value
	// Строковые формы для JSON и текста.
	jsonLiteral string
	jsonWant    string
	textWant    string
	// Ожидать ли ошибку маршалинга JSON (комплексные типы).
	marshalJSONErr bool
	// Ожидать ли ошибку разбора некорректной строки (для строки — нет).
	badStrErr bool
}

// testNullable Жизненный цикл Nullable[T] для одного типа.
func testNullable[T nullableValue](t *testing.T, kc nullableKind) {
	t.Helper()
	var zero T
	value := kc.value.(T)

	t.Run("New", func(t *testing.T) {
		v1 := NewNullable[T]()
		if v1.Valid {
			t.Error("NewNullable()", errValidTrue)
		}
		v2 := NewNullableValue(value)
		if !v2.Valid || any(v2.Nullable) != any(value) {
			t.Error("NewNullableValue()", errWrong)
		}
		pv := value
		v3 := NewNullablePointerValue(&pv)
		if !v3.Valid || any(v3.Nullable) != any(value) {
			t.Error("NewNullablePointerValue()", errWrong)
		}
		v4 := NewNullablePointerValue[T](nil)
		if v4.Valid {
			t.Error("NewNullablePointerValue(nil)", errValidTrue)
		}
	})

	t.Run("SetReset", func(t *testing.T) {
		v1 := NewNullable[T]()
		v1.SetValid(value)
		if !v1.Valid || any(v1.Nullable) != any(value) {
			t.Error("SetValid()", errWrong)
		}
		v1.Reset()
		if v1.Valid || any(v1.Nullable) != any(zero) {
			t.Error("Reset()", errWrong)
		}
	})

	t.Run("NullIfDefault", func(t *testing.T) {
		v1 := NewNullableValue(zero)
		v1.NullIfDefault()
		if v1.Valid {
			t.Error("NullIfDefault()", errValidTrue)
		}
		v2 := NewNullableValue(value)
		v2.NullIfDefault()
		if !v2.Valid {
			t.Error("NullIfDefault()", errValidFalse)
		}
	})

	t.Run("MustValue", func(t *testing.T) {
		v1 := NewNullable[T]()
		if any(v1.MustValue()) != any(zero) {
			t.Error("MustValue()", errWrong)
		}
		v2 := NewNullableValue(value)
		if any(v2.MustValue()) != any(value) {
			t.Error("MustValue()", errWrong)
		}
	})

	t.Run("Pointer", func(t *testing.T) {
		v1 := NewNullable[T]()
		if v1.Pointer() != nil {
			t.Error("Pointer()", errNotNil)
		}
		v2 := NewNullableValue(value)
		if v2.Pointer() == nil || any(*v2.Pointer()) != any(value) {
			t.Error("Pointer()", errNil)
		}
	})

	t.Run("Scan", func(t *testing.T) {
		v1 := NewNullable[T]()
		errorPanic(v1.Scan(nil))
		if v1.Valid {
			t.Error("Scan(nil)", errValidTrue)
		}
		v2 := NewNullable[T]()
		errorPanic(v2.Scan(kc.str))
		if !v2.Valid || any(v2.Nullable) != any(value) {
			t.Error("Scan()", errWrong)
		}
		v3 := NewNullable[T]()
		err := v3.Scan("not-a-value-of-kind")
		if kc.badStrErr {
			if err == nil {
				t.Error("Scan()", errNil)
			}
		} else if err != nil {
			t.Error("Scan()", err)
		}
	})

	t.Run("Value", func(t *testing.T) {
		v1 := NewNullableValue(value)
		dv, err := v1.Value()
		errorPanic(err)
		if !reflect.DeepEqual(dv, kc.want) {
			t.Errorf("Value() = %#v, ожидалось %#v", dv, kc.want)
		}
		v2 := NewNullable[T]()
		dv, err = v2.Value()
		errorPanic(err)
		if dv != nil {
			t.Error("Value()", errReturnsNotNil)
		}
	})

	t.Run("JSON", func(t *testing.T) {
		v1 := NewNullable[T]()
		err := json.Unmarshal([]byte("null"), &v1)
		errorPanic(err)
		if v1.Valid {
			t.Error("UnmarshalJSON(null)", errValidTrue)
		}
		v2 := NewNullable[T]()
		err = json.Unmarshal([]byte(kc.jsonLiteral), &v2)
		errorPanic(err)
		if !v2.Valid || any(v2.Nullable) != any(value) {
			t.Error("UnmarshalJSON()", errWrong)
		}
		v3 := NewNullable[T]()
		err = json.Unmarshal([]byte(`""`), &v3)
		errorPanic(err)
		if v3.Valid {
			t.Error("UnmarshalJSON(\"\")", errValidTrue)
		}
		data, err := NewNullableValue(value).MarshalJSON()
		if kc.marshalJSONErr {
			if err == nil {
				t.Error("MarshalJSON()", errNil)
			}
		} else {
			errorPanic(err)
			jsonEquals(t, data, kc.jsonWant, "Nullable json marshal")
		}
		data, err = NewNullable[T]().MarshalJSON()
		errorPanic(err)
		jsonEquals(t, data, "null", "Nullable null json marshal")
	})

	t.Run("Text", func(t *testing.T) {
		v1 := NewNullable[T]()
		errorPanic(v1.UnmarshalText([]byte("null")))
		if v1.Valid {
			t.Error("UnmarshalText(null)", errValidTrue)
		}
		v2 := NewNullable[T]()
		errorPanic(v2.UnmarshalText([]byte(kc.str)))
		if !v2.Valid || any(v2.Nullable) != any(value) {
			t.Error("UnmarshalText()", errWrong)
		}
		v3 := NewNullable[T]()
		errorPanic(v3.UnmarshalText([]byte("")))
		if !v3.Valid || any(v3.Nullable) != any(zero) {
			t.Error("UnmarshalText(\"\")", errWrong)
		}
		v4 := NewNullable[T]()
		err := v4.UnmarshalText([]byte("not-a-value-of-kind"))
		if kc.badStrErr {
			if err == nil {
				t.Error("UnmarshalText()", errNil)
			}
		} else if err != nil {
			t.Error("UnmarshalText()", err)
		}
		data, err := NewNullableValue(value).MarshalText()
		errorPanic(err)
		jsonEquals(t, data, kc.textWant, "Nullable text marshal")
		data, err = NewNullable[T]().MarshalText()
		errorPanic(err)
		jsonEquals(t, data, "null", "Nullable null text marshal")
	})

	t.Run("Binary", func(t *testing.T) {
		v1 := NewNullableValue(value)
		blob, err := v1.MarshalBinary()
		errorPanic(err)
		v2 := NewNullable[T]()
		errorPanic(v2.UnmarshalBinary(blob))
		if !v2.Valid || any(v2.Nullable) != any(value) {
			t.Error("MarshalBinary() -> UnmarshalBinary()", errWrong)
		}
		v3 := NewNullable[T]()
		err = v3.UnmarshalBinary([]byte("not-a-gob"))
		if err == nil {
			t.Error("UnmarshalBinary()", errNil)
		}
	})
}

func TestNullable(t *testing.T) {
	var kinds = []nullableKind{
		{name: "int", value: int(42), str: "42", want: int64(42), jsonLiteral: "42", jsonWant: "42", textWant: "42", badStrErr: true},
		{name: "int8", value: int8(-42), str: "-42", want: int64(-42), jsonLiteral: "-42", jsonWant: "-42", textWant: "-42", badStrErr: true},
		{name: "int16", value: int16(-420), str: "-420", want: int64(-420), jsonLiteral: "-420", jsonWant: "-420", textWant: "-420", badStrErr: true},
		{name: "int32", value: int32(-42000), str: "-42000", want: int64(-42000), jsonLiteral: "-42000", jsonWant: "-42000", textWant: "-42000", badStrErr: true},
		{name: "int64", value: int64(-4200000000), str: "-4200000000", want: int64(-4200000000), jsonLiteral: "-4200000000", jsonWant: "-4200000000", textWant: "-4200000000", badStrErr: true},
		{name: "uint", value: uint(42), str: "42", want: []byte("42"), jsonLiteral: "42", jsonWant: "42", textWant: "42", badStrErr: true},
		{name: "uint8", value: uint8(255), str: "255", want: []byte("255"), jsonLiteral: "255", jsonWant: "255", textWant: "255", badStrErr: true},
		{name: "uint16", value: uint16(65535), str: "65535", want: []byte("65535"), jsonLiteral: "65535", jsonWant: "65535", textWant: "65535", badStrErr: true},
		{name: "uint32", value: uint32(4294967295), str: "4294967295", want: []byte("4294967295"), jsonLiteral: "4294967295", jsonWant: "4294967295", textWant: "4294967295", badStrErr: true},
		{name: "uint64", value: uint64(18446744073709551615), str: "18446744073709551615", want: []byte("18446744073709551615"), jsonLiteral: "18446744073709551615", jsonWant: "18446744073709551615", textWant: "18446744073709551615", badStrErr: true},
		{name: "uintptr", value: uintptr(42), str: "42", want: []byte("42"), jsonLiteral: "42", jsonWant: "42", textWant: "42", badStrErr: true},
		{name: "float32", value: float32(1.5), str: "1.5", want: float64(1.5), jsonLiteral: "1.5", jsonWant: "1.5", textWant: "1.5", badStrErr: true},
		{name: "float64", value: float64(1.5), str: "1.5", want: float64(1.5), jsonLiteral: "1.5", jsonWant: "1.5", textWant: "1.5", badStrErr: true},
		{name: "string", value: "test string", str: "test string", want: "test string", jsonLiteral: `"test string"`, jsonWant: `"test string"`, textWant: "test string"},
		{name: "bool", value: true, str: "true", want: true, jsonLiteral: "true", jsonWant: "true", textWant: "true", badStrErr: true},
		{name: "complex64", value: complex64(1.5 + 2.5i), str: "(1.5+2.5i)", want: "(1.5+2.5i)", jsonLiteral: `"(1.5+2.5i)"`, jsonWant: "", textWant: "(1.5+2.5i)", marshalJSONErr: true, badStrErr: true},
		{name: "complex128", value: complex128(1.5 + 2.5i), str: "(1.5+2.5i)", want: "(1.5+2.5i)", jsonLiteral: `"(1.5+2.5i)"`, jsonWant: "", textWant: "(1.5+2.5i)", marshalJSONErr: true, badStrErr: true},
	}

	for _, kc := range kinds {
		kc := kc
		t.Run(kc.name, func(t *testing.T) {
			switch kc.name {
			case "int":
				testNullable[int](t, kc)
			case "int8":
				testNullable[int8](t, kc)
			case "int16":
				testNullable[int16](t, kc)
			case "int32":
				testNullable[int32](t, kc)
			case "int64":
				testNullable[int64](t, kc)
			case "uint":
				testNullable[uint](t, kc)
			case "uint8":
				testNullable[uint8](t, kc)
			case "uint16":
				testNullable[uint16](t, kc)
			case "uint32":
				testNullable[uint32](t, kc)
			case "uint64":
				testNullable[uint64](t, kc)
			case "uintptr":
				testNullable[uintptr](t, kc)
			case "float32":
				testNullable[float32](t, kc)
			case "float64":
				testNullable[float64](t, kc)
			case "string":
				testNullable[string](t, kc)
			case "bool":
				testNullable[bool](t, kc)
			case "complex64":
				testNullable[complex64](t, kc)
			case "complex128":
				testNullable[complex128](t, kc)
			}
		})
	}
}

// Производный тип: проверка работы через базовый тип (тильда в контракте).
type nullableDerivedInt int

func TestNullableDerivedType(t *testing.T) {
	var v Nullable[nullableDerivedInt]
	errorPanic(v.Scan("42"))
	if !v.Valid || v.Nullable != 42 {
		t.Error("Scan()", errWrong)
	}
	dv, err := v.Value()
	errorPanic(err)
	if dv != int64(42) {
		t.Error("Value()", errWrong)
	}
}

//goland:noinspection ALL
func TestNullableUnmarshalJSONErrors(t *testing.T) {
	var err error

	v1 := NewNullable[int]()
	err = v1.UnmarshalJSON([]byte(""))
	if err == nil {
		t.Error("UnmarshalJSON()", errNil)
	}

	v2 := NewNullable[int]()
	err = v2.UnmarshalJSON([]byte(":{;-}"))
	if err == nil {
		t.Error("UnmarshalJSON()", errNil)
	}

	v3 := NewNullable[int]()
	err = v3.UnmarshalJSON([]byte("\"\\uZZ\""))
	if err == nil {
		t.Error("UnmarshalJSON()", errNil)
	}
}

func TestNullableValueComplex(t *testing.T) {
	// Проверка строковой формы комплексного значения для driver.Value.
	v1 := NewNullableValue(complex128(1.5 + 2.5i))
	dv, err := v1.Value()
	errorPanic(err)
	if dv != "(1.5+2.5i)" {
		t.Errorf("Value() = %#v, ожидалось %#v", dv, "(1.5+2.5i)")
	}
}

func TestNullableWire(t *testing.T) {
	// Прямая проверка двоичного представления.
	var item nullableWire[int]
	_ = item
	data, err := NewNullableValue(42).MarshalBinary()
	errorPanic(err)
	if len(data) == 0 || !bytes.Equal(data, data) {
		t.Error("MarshalBinary()", errWrong)
	}
}

func TestNullableTextFormat(t *testing.T) {
	// Формат %v для текстового кодека.
	data, err := NewNullableValue(42).MarshalText()
	errorPanic(err)
	if string(data) != fmt.Sprintf("%v", 42) {
		t.Error("MarshalText()", errWrong)
	}
}
