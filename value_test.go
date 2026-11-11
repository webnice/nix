package nix

import (
	"encoding/json"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"
)

// ptrOf Возвращает указатель на значение для помещения в интерфейс.
func ptrOf[T any](v T) *T { return &v }

// TestValueInterface Операции интерфейса Interface над гетерогенной коллекцией
// значений всех типов пакета.
func TestValueInterface(t *testing.T) {
	values := []Interface{
		&Bool{}, ptrOf(NewBytes()), &Complex128{}, &Complex64{}, &Duration{}, &Float32{}, &Float64{},
		&Int{}, &Int16{}, &Int32{}, &Int64{}, &Int8{}, &IP{}, &RawMessage{}, &String{}, &StringSlice{},
		&Time{}, &Uint{}, &Uint16{}, &Uint32{}, &Uint64{}, &Uint8{}, &Uintptr{}, &UUID{},
		ptrOf(NewBigInt()), &Nullable[int]{},
	}

	for _, v := range values {
		// Сброс и сканирование NULL.
		v.Reset()
		if err := v.Scan(nil); err != nil {
			t.Errorf("Scan(nil): %v", err)
		}
		dv, err := v.Value()
		if err != nil {
			t.Errorf("Value(): %v", err)
		}
		if dv != nil {
			t.Errorf("Value() после Scan(nil) = %#v, ожидалось nil", dv)
		}
		// Сериализация недействительного значения.
		if data, err := v.MarshalJSON(); err != nil || string(data) != "null" {
			t.Errorf("MarshalJSON() = %q, %v; ожидалось null", data, err)
		}
		if data, err := v.MarshalText(); err != nil || string(data) != "null" {
			t.Errorf("MarshalText() = %q, %v; ожидалось null", data, err)
		}
		blob, err := v.MarshalBinary()
		if err != nil || len(blob) == 0 {
			t.Errorf("MarshalBinary() = %d байт, %v", len(blob), err)
		}
		// Двоичный цикл через интерфейс.
		if err = v.UnmarshalBinary(blob); err != nil {
			t.Errorf("UnmarshalBinary(): %v", err)
		}
		// Разбор JSON null и текста null.
		if err = v.UnmarshalJSON([]byte("null")); err != nil {
			t.Errorf("UnmarshalJSON(null): %v", err)
		}
		if err = v.UnmarshalText([]byte("null")); err != nil {
			t.Errorf("UnmarshalText(null): %v", err)
		}
	}
}

// valuerTest Типизированные операции интерфейса Valuer[T].
func valuerTest[T any](t *testing.T, v Valuer[T], value T) {
	t.Helper()

	// Set присваивает значение как есть.
	v.Set(value)
	if dv, err := v.Value(); err != nil || dv == nil {
		t.Errorf("Set(): Value() = %#v, %v", dv, err)
	}
	if got := v.MustValue(); !reflect.DeepEqual(got, value) {
		t.Errorf("MustValue() = %#v, ожидалось %#v", got, value)
	}
	if p := v.Pointer(); p == nil || !reflect.DeepEqual(*p, value) {
		t.Errorf("Pointer() = %#v, ожидалось %#v", p, value)
	}

	// SetValid присваивает значение, не являющееся дефолтным.
	v.SetValid(value)
	if got := v.MustValue(); !reflect.DeepEqual(got, value) {
		t.Errorf("SetValid(): MustValue() = %#v, ожидалось %#v", got, value)
	}

	// Reset сбрасывает значение: Pointer() единообразно возвращает nil.
	v.Reset()
	if p := v.Pointer(); p != nil {
		t.Errorf("Pointer() после Reset() = %#v, ожидался nil", p)
	}
}

// TestValuerInterface Типизированная работа через интерфейс Valuer[T].
func TestValuerInterface(t *testing.T) {
	t.Run("BigInt", func(t *testing.T) {
		valuerTest[*big.Int](t, &BigInt{}, bigIntOf("18446744073709551616"))
	})
	t.Run("Bool", func(t *testing.T) { valuerTest[bool](t, &Bool{}, true) })
	t.Run("Bytes", func(t *testing.T) { valuerTest[[]byte](t, &Bytes{}, []byte("test")) })
	t.Run("Complex64", func(t *testing.T) { valuerTest[complex64](t, &Complex64{}, complex64(1.5+2.5i)) })
	t.Run("Complex128", func(t *testing.T) { valuerTest[complex128](t, &Complex128{}, complex128(1.5+2.5i)) })
	t.Run("Duration", func(t *testing.T) { valuerTest[time.Duration](t, &Duration{}, 90*time.Minute) })
	t.Run("Float32", func(t *testing.T) { valuerTest[float32](t, &Float32{}, 1.5) })
	t.Run("Float64", func(t *testing.T) { valuerTest[float64](t, &Float64{}, 1.5) })
	t.Run("Int", func(t *testing.T) { valuerTest[int](t, &Int{}, 42) })
	t.Run("Int8", func(t *testing.T) { valuerTest[int8](t, &Int8{}, -42) })
	t.Run("Int16", func(t *testing.T) { valuerTest[int16](t, &Int16{}, -420) })
	t.Run("Int32", func(t *testing.T) { valuerTest[int32](t, &Int32{}, -42000) })
	t.Run("Int64", func(t *testing.T) { valuerTest[int64](t, &Int64{}, -4200000000) })
	t.Run("IP", func(t *testing.T) { valuerTest[net.IP](t, &IP{}, net.ParseIP("192.168.0.1")) })
	t.Run("RawMessage", func(t *testing.T) {
		valuerTest[json.RawMessage](t, &RawMessage{}, json.RawMessage(`{"key":"value"}`))
	})
	t.Run("String", func(t *testing.T) { valuerTest[string](t, &String{}, "test") })
	t.Run("StringSlice", func(t *testing.T) { valuerTest[[]string](t, &StringSlice{}, []string{"one", "two"}) })
	t.Run("Time", func(t *testing.T) { valuerTest[time.Time](t, &Time{}, timeOkValidValue) })
	t.Run("Uint", func(t *testing.T) { valuerTest[uint](t, &Uint{}, 42) })
	t.Run("Uint8", func(t *testing.T) { valuerTest[uint8](t, &Uint8{}, 255) })
	t.Run("Uint16", func(t *testing.T) { valuerTest[uint16](t, &Uint16{}, 65535) })
	t.Run("Uint32", func(t *testing.T) { valuerTest[uint32](t, &Uint32{}, 4294967295) })
	t.Run("Uint64", func(t *testing.T) { valuerTest[uint64](t, &Uint64{}, 18446744073709551615) })
	t.Run("Uintptr", func(t *testing.T) { valuerTest[uintptr](t, &Uintptr{}, 42) })
	t.Run("UUID", func(t *testing.T) {
		valuerTest[[16]byte](t, &UUID{}, [16]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff})
	})
	t.Run("Nullable", func(t *testing.T) { valuerTest[int](t, &Nullable[int]{}, 42) })
}

// validOf Возвращает значение поля Valid обёртки, вспомогательная функция теста.
func validOf(v any) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.FieldByName("Valid").Bool()
}

// nullifierTest Операции пересечения интерфейсов Valuer[T] и Nullifier[W].
func nullifierTest[T any, W any](t *testing.T, v interface {
	Valuer[T]
	Nullifier[W]
}, value T, zero T) {
	t.Helper()

	// Недефолтное значение остаётся действительным.
	v.Set(value)
	if got := v.NullIfDefault(); !validOf(got) {
		t.Error("NullIfDefault() недефолтного значения: Valid=false")
	}
	// Дефолтное значение сбрасывается до nil.
	v.Set(zero)
	if got := v.NullIfDefault(); validOf(got) {
		t.Error("NullIfDefault() дефолтного значения: Valid=true")
	}
}

// TestNullifierInterface Работа через пересечение Valuer[T] и Nullifier[W].
func TestNullifierInterface(t *testing.T) {
	t.Run("BigInt", func(t *testing.T) {
		nullifierTest[*big.Int, BigInt](t, &BigInt{}, bigIntOf("18446744073709551616"), nil)
	})
	t.Run("Bool", func(t *testing.T) { nullifierTest[bool, Bool](t, &Bool{}, true, false) })
	t.Run("Bytes", func(t *testing.T) { nullifierTest[[]byte, Bytes](t, &Bytes{}, []byte("test"), []byte{}) })
	t.Run("Complex64", func(t *testing.T) {
		nullifierTest[complex64, Complex64](t, &Complex64{}, complex64(1.5+2.5i), 0)
	})
	t.Run("Complex128", func(t *testing.T) {
		nullifierTest[complex128, Complex128](t, &Complex128{}, complex128(1.5+2.5i), 0)
	})
	t.Run("Duration", func(t *testing.T) {
		nullifierTest[time.Duration, Duration](t, &Duration{}, 90*time.Minute, 0)
	})
	t.Run("Float32", func(t *testing.T) { nullifierTest[float32, Float32](t, &Float32{}, 1.5, 0) })
	t.Run("Float64", func(t *testing.T) { nullifierTest[float64, Float64](t, &Float64{}, 1.5, 0) })
	t.Run("Int", func(t *testing.T) { nullifierTest[int, Int](t, &Int{}, 42, 0) })
	t.Run("Int8", func(t *testing.T) { nullifierTest[int8, Int8](t, &Int8{}, -42, 0) })
	t.Run("Int16", func(t *testing.T) { nullifierTest[int16, Int16](t, &Int16{}, -420, 0) })
	t.Run("Int32", func(t *testing.T) { nullifierTest[int32, Int32](t, &Int32{}, -42000, 0) })
	t.Run("Int64", func(t *testing.T) { nullifierTest[int64, Int64](t, &Int64{}, -4200000000, 0) })
	t.Run("IP", func(t *testing.T) { nullifierTest[net.IP, IP](t, &IP{}, net.ParseIP("192.168.0.1"), nil) })
	t.Run("RawMessage", func(t *testing.T) {
		nullifierTest[json.RawMessage, RawMessage](t, &RawMessage{}, json.RawMessage(`{"key":"value"}`), nil)
	})
	t.Run("String", func(t *testing.T) { nullifierTest[string, String](t, &String{}, "test", "") })
	t.Run("StringSlice", func(t *testing.T) {
		nullifierTest[[]string, StringSlice](t, &StringSlice{}, []string{"test"}, []string{})
	})
	t.Run("Time", func(t *testing.T) { nullifierTest[time.Time, Time](t, &Time{}, timeOkValidValue, time.Time{}) })
	t.Run("Uint", func(t *testing.T) { nullifierTest[uint, Uint](t, &Uint{}, 42, 0) })
	t.Run("Uint8", func(t *testing.T) { nullifierTest[uint8, Uint8](t, &Uint8{}, 255, 0) })
	t.Run("Uint16", func(t *testing.T) { nullifierTest[uint16, Uint16](t, &Uint16{}, 65535, 0) })
	t.Run("Uint32", func(t *testing.T) { nullifierTest[uint32, Uint32](t, &Uint32{}, 4294967295, 0) })
	t.Run("Uint64", func(t *testing.T) {
		nullifierTest[uint64, Uint64](t, &Uint64{}, 18446744073709551615, 0)
	})
	t.Run("Uintptr", func(t *testing.T) { nullifierTest[uintptr, Uintptr](t, &Uintptr{}, 42, 0) })
	t.Run("UUID", func(t *testing.T) {
		nullifierTest[[16]byte, UUID](t, &UUID{}, [16]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}, [16]byte{})
	})
	t.Run("Nullable", func(t *testing.T) { nullifierTest[int, Nullable[int]](t, &Nullable[int]{}, 42, 0) })
}

// TestValueInterfaceCompileTime Проверка удовлетворения интерфейсов в рантайме.
func TestValueInterfaceRuntime(t *testing.T) {
	var anyv any = &Int64{}
	if _, ok := anyv.(Interface); !ok {
		t.Error("*Int64 не реализует Interface")
	}
	if _, ok := anyv.(Valuer[int64]); !ok {
		t.Error("*Int64 не реализует Valuer[int64]")
	}
	// Значение (не указатель) Value не реализует: методы изменения на указателе.
	anyv = Int64{}
	if _, ok := anyv.(Interface); ok {
		t.Error("Int64 (значение) не должен реализовывать Interface")
	}
}
