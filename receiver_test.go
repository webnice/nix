package nix

// Регрессионные проверки приёмников методов.
//
// Методы Value(), MarshalJSON(), MarshalText() и MarshalBinary() обязаны быть объявлены
// на значении типа (value receiver), а не на указателе. Типы используются как поля
// структур, и ORM (GORM, sqlx и другие) передают значения этих полей в database/sql
// как есть — значением, а не указателем. Если перенести методы на *T, значение типа
// перестанет удовлетворять интерфейсам driver.Valuer, json.Marshaler,
// encoding.TextMarshaler и encoding.BinaryMarshaler: запись в БД завершится ошибкой
// «unsupported type nix.X, a struct», а сериализация будет выдавать объект
// {«Поле»:…,«Valid»:…} вместо значения.
// Менять приёмник этих методов на указатель ЗАПРЕЩЕНО — проверки в этом файле
// обязаны оставаться зелёными.

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"math"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"
)

// bigIntOf Действительный объект BigInt из строки, вспомогательная функция теста.
func bigIntOf(s string) *big.Int {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid big.Int string")
	}
	return v
}

// uuidOf Значение UUID из строки, вспомогательная функция теста.
func uuidOf(s string) (u [16]byte) {
	var err error
	if u, err = uuidParse(s); err != nil {
		panic(err)
	}
	return
}

// Проверки на этапе компиляции: значение каждого типа, без взятия адреса,
// должно удовлетворять интерфейсам, используемым database/sql и кодеками.
// При смене приёмника методов на *T компиляция тестов пакета сломается сразу.
var (
	_ driver.Valuer = Bool{}
	_ driver.Valuer = Bytes{}
	_ driver.Valuer = Float64{}
	_ driver.Valuer = Int64{}
	_ driver.Valuer = String{}
	_ driver.Valuer = StringSlice{}
	_ driver.Valuer = Time{}
	_ driver.Valuer = Uint64{}
	_ driver.Valuer = Int{}
	_ driver.Valuer = Int8{}
	_ driver.Valuer = Int16{}
	_ driver.Valuer = Int32{}
	_ driver.Valuer = Uint{}
	_ driver.Valuer = Uint8{}
	_ driver.Valuer = Uint16{}
	_ driver.Valuer = Uint32{}
	_ driver.Valuer = Uintptr{}
	_ driver.Valuer = Float32{}
	_ driver.Valuer = Complex64{}
	_ driver.Valuer = Complex128{}
	_ driver.Valuer = Duration{}
	_ driver.Valuer = RawMessage{}
	_ driver.Valuer = IP{}
	_ driver.Valuer = BigInt{}
	_ driver.Valuer = UUID{}
	_ driver.Valuer = Nullable[int]{}

	_ json.Marshaler = Bool{}
	_ json.Marshaler = Bytes{}
	_ json.Marshaler = Float64{}
	_ json.Marshaler = Int64{}
	_ json.Marshaler = String{}
	_ json.Marshaler = StringSlice{}
	_ json.Marshaler = Time{}
	_ json.Marshaler = Uint64{}
	_ json.Marshaler = Int{}
	_ json.Marshaler = Int8{}
	_ json.Marshaler = Int16{}
	_ json.Marshaler = Int32{}
	_ json.Marshaler = Uint{}
	_ json.Marshaler = Uint8{}
	_ json.Marshaler = Uint16{}
	_ json.Marshaler = Uint32{}
	_ json.Marshaler = Uintptr{}
	_ json.Marshaler = Float32{}
	_ json.Marshaler = Complex64{}
	_ json.Marshaler = Complex128{}
	_ json.Marshaler = Duration{}
	_ json.Marshaler = RawMessage{}
	_ json.Marshaler = IP{}
	_ json.Marshaler = BigInt{}
	_ json.Marshaler = UUID{}
	_ json.Marshaler = Nullable[int]{}

	_ encoding.TextMarshaler = Bool{}
	_ encoding.TextMarshaler = Bytes{}
	_ encoding.TextMarshaler = Float64{}
	_ encoding.TextMarshaler = Int64{}
	_ encoding.TextMarshaler = String{}
	_ encoding.TextMarshaler = StringSlice{}
	_ encoding.TextMarshaler = Time{}
	_ encoding.TextMarshaler = Uint64{}
	_ encoding.TextMarshaler = Int{}
	_ encoding.TextMarshaler = Int8{}
	_ encoding.TextMarshaler = Int16{}
	_ encoding.TextMarshaler = Int32{}
	_ encoding.TextMarshaler = Uint{}
	_ encoding.TextMarshaler = Uint8{}
	_ encoding.TextMarshaler = Uint16{}
	_ encoding.TextMarshaler = Uint32{}
	_ encoding.TextMarshaler = Uintptr{}
	_ encoding.TextMarshaler = Float32{}
	_ encoding.TextMarshaler = Complex64{}
	_ encoding.TextMarshaler = Complex128{}
	_ encoding.TextMarshaler = Duration{}
	_ encoding.TextMarshaler = RawMessage{}
	_ encoding.TextMarshaler = IP{}
	_ encoding.TextMarshaler = BigInt{}
	_ encoding.TextMarshaler = UUID{}
	_ encoding.TextMarshaler = Nullable[int]{}

	_ encoding.BinaryMarshaler = Bool{}
	_ encoding.BinaryMarshaler = Bytes{}
	_ encoding.BinaryMarshaler = Float64{}
	_ encoding.BinaryMarshaler = Int64{}
	_ encoding.BinaryMarshaler = String{}
	_ encoding.BinaryMarshaler = StringSlice{}
	_ encoding.BinaryMarshaler = Time{}
	_ encoding.BinaryMarshaler = Uint64{}
	_ encoding.BinaryMarshaler = Int{}
	_ encoding.BinaryMarshaler = Int8{}
	_ encoding.BinaryMarshaler = Int16{}
	_ encoding.BinaryMarshaler = Int32{}
	_ encoding.BinaryMarshaler = Uint{}
	_ encoding.BinaryMarshaler = Uint8{}
	_ encoding.BinaryMarshaler = Uint16{}
	_ encoding.BinaryMarshaler = Uint32{}
	_ encoding.BinaryMarshaler = Uintptr{}
	_ encoding.BinaryMarshaler = Float32{}
	_ encoding.BinaryMarshaler = Complex64{}
	_ encoding.BinaryMarshaler = Complex128{}
	_ encoding.BinaryMarshaler = Duration{}
	_ encoding.BinaryMarshaler = RawMessage{}
	_ encoding.BinaryMarshaler = IP{}
	_ encoding.BinaryMarshaler = BigInt{}
	_ encoding.BinaryMarshaler = UUID{}
	_ encoding.BinaryMarshaler = Nullable[int]{}
)

// receiverTestLastArgs Аргументы последнего выполненного запроса, полученные драйвером.
var receiverTestLastArgs []driver.Value

// receiverTestDriver Фейковый драйвер БД: фиксирует аргументы запроса.
type receiverTestDriver struct{}

func (receiverTestDriver) Open(string) (driver.Conn, error) { return &receiverTestConn{}, nil }

type receiverTestConn struct{}

func (c *receiverTestConn) Prepare(string) (driver.Stmt, error) { return &receiverTestStmt{}, nil }
func (c *receiverTestConn) Close() error                        { return nil }
func (c *receiverTestConn) Begin() (driver.Tx, error)           { return &receiverTestTx{}, nil }

type receiverTestTx struct{}

func (receiverTestTx) Commit() error   { return nil }
func (receiverTestTx) Rollback() error { return nil }

type receiverTestStmt struct{}

func (s *receiverTestStmt) Close() error  { return nil }
func (s *receiverTestStmt) NumInput() int { return -1 }

func (s *receiverTestStmt) Exec(args []driver.Value) (driver.Result, error) {
	receiverTestLastArgs = append([]driver.Value(nil), args...)
	return driver.RowsAffected(1), nil
}

func (s *receiverTestStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, driver.ErrSkip
}

func init() { sql.Register("nix-receiver-test", receiverTestDriver{}) }

// TestValueReceiverDatabaseSQL Поля структуры со значениями типов пакета передаются
// в database/sql значениями (не указателями) — именно так их передают ORM при
// INSERT и UPDATE. При смене приёмника Value() на *T этот тест падает с ошибкой
// «sql: converting argument $N type: unsupported type nix.X, a struct».
func TestValueReceiverDatabaseSQL(t *testing.T) {
	var (
		db  *sql.DB
		err error
	)

	if db, err = sql.Open("nix-receiver-test", ""); err != nil {
		t.Fatalf("sql.Open() ошибка: %v", err)
	}
	defer func() { _ = db.Close() }()

	// message Структура-аналог моделей с полями типов пакета и тегами db/gorm.
	type message struct {
		Bool       Bool          `db:"bool"     gorm:"column:bool"`
		Bytes      Bytes         `db:"bytes"    gorm:"column:bytes"`
		Float64    Float64       `db:"float64"  gorm:"column:float64"`
		Int64      Int64         `db:"int64"    gorm:"column:int64"`
		String     String        `db:"string"   gorm:"column:string"`
		StringSl   StringSlice   `db:"stringsl" gorm:"column:stringsl"`
		Time       Time          `db:"time"     gorm:"column:time"`
		Uint64     Uint64        `db:"uint64"     gorm:"column:uint64"`
		Duration   Duration      `db:"duration"   gorm:"column:duration"`
		RawMessage RawMessage    `db:"dataRaw"    gorm:"column:dataRaw"`
		IP         IP            `db:"ip"         gorm:"column:ip"`
		BigInt     BigInt        `db:"bigInt"     gorm:"column:bigInt"`
		UUID       UUID          `db:"uuid"       gorm:"column:uuid"`
		Complex    Complex64     `db:"complex"    gorm:"column:complex"`
		Int        Int           `db:"int"        gorm:"column:int"`
		Nullable   Nullable[int] `db:"nullable" gorm:"column:nullable"`
	}

	msg := message{
		Bool:       NewBoolValue(true),
		Bytes:      NewBytesValue([]byte("test-bytes")),
		Float64:    NewFloat64Value(1.5),
		Int64:      NewInt64Value(-42),
		String:     NewStringValue("test-string"),
		StringSl:   NewStringSliceValue([]string{"test-one", "test-two"}),
		Time:       NewTimeValue(timeOkValidValue),
		Uint64:     NewUint64Value(math.MaxUint64),
		Duration:   NewDurationValue(90 * time.Minute),
		RawMessage: NewRawMessageValue([]byte(`{"key":"value"}`)),
		IP:         NewIPValue(net.ParseIP("192.168.0.1")),
		BigInt:     NewBigIntValue(bigIntOf("18446744073709551616")),
		UUID:       NewUUIDValue(uuidOf("00112233-4455-6677-8899-aabbccddeeff")),
		Complex:    NewComplex64Value(complex64(1.5 + 2.5i)),
		Int:        NewIntValue(42),
		Nullable:   NewNullableValue(42),
	}

	if _, err = db.Exec(
		"INSERT INTO message VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		msg.Bool, msg.Bytes, msg.Float64, msg.Int64, msg.String, msg.StringSl, msg.Time, msg.Uint64,
		msg.Duration, msg.RawMessage, msg.IP, msg.BigInt, msg.UUID, msg.Complex, msg.Int, msg.Nullable,
	); err != nil {
		t.Fatalf("Передача значений полей структуры в database/sql: %v", err)
	}

	//goland:noinspection GoRedundantConversion
	want := []driver.Value{
		true,
		[]byte("test-bytes"),
		float64(1.5),
		int64(-42),
		"test-string",
		"test-one,test-two",
		timeOkValidValue,
		[]byte("18446744073709551615"),
		int64(90 * time.Minute),
		[]byte(`{"key":"value"}`),
		"192.168.0.1",
		"18446744073709551616",
		"00112233-4455-6677-8899-aabbccddeeff",
		"(1.5+2.5i)",
		int64(42),
		int64(42),
	}
	if !reflect.DeepEqual(receiverTestLastArgs, want) {
		t.Errorf("Драйвер получил аргументы: %#v, ожидалось: %#v", receiverTestLastArgs, want)
	}
}

// TestValueReceiverJSON Сериализация не адресуемых значений (переданных в json.Marshal
// как any, то есть без возможности взять адрес) обязана давать значение,
// а не объект вида {"Поле":…,"Valid":…}. При смене приёмника MarshalJSON() на *T
// этот тест падает: значение перестаёт реализовать json.Marshaler.
func TestValueReceiverJSON(t *testing.T) {
	var cases = []struct {
		name  string
		value any
		want  []byte
	}{
		{"Bool", NewBoolValue(true), []byte(`true`)},
		{"Bool null", NewBool(), []byte(`null`)},
		{"Bytes", NewBytesValue([]byte("test-bytes")), []byte(`"dGVzdC1ieXRlcw=="`)},
		{"Float64", NewFloat64Value(math.MaxFloat64), float64JSON},
		{"Int64", NewInt64Value(math.MaxInt64), int64JSON},
		{"String", NewStringValue(stringTestBody), stringJSON},
		{"StringSlice", NewStringSliceValue([]string{"one", "two"}), []byte(`["one","two"]`)},
		{"Time", NewTimeValue(timeOkValidValue), timeStringValueJSON},
		{"Uint64", NewUint64Value(math.MaxUint64), uint64JSON},
		{"Int", NewIntValue(-42), []byte(`-42`)},
		{"Uint8 null", NewUint8(), []byte(`null`)},
		{"Float32", NewFloat32Value(1.5), []byte(`1.500000`)},
		{"Complex64", NewComplex64Value(complex64(1.5 + 2.5i)), []byte(`"(1.5+2.5i)"`)},
		{"Duration", NewDurationValue(90 * time.Minute), []byte(`5400000000000`)},
		{"RawMessage", NewRawMessageValue([]byte(`{"key":"value"}`)), []byte(`{"key":"value"}`)},
		{"IP", NewIPValue(net.ParseIP("192.168.0.1")), []byte(`"192.168.0.1"`)},
		{"BigInt", NewBigIntValue(bigIntOf("18446744073709551616")), []byte(`18446744073709551616`)},
		{"UUID", NewUUIDValue(uuidOf("00112233-4455-6677-8899-aabbccddeeff")), []byte(`"00112233-4455-6677-8899-aabbccddeeff"`)},
		{"Nullable", NewNullableValue(42), []byte(`42`)},
		{"Nullable null", NewNullable[int](), []byte(`null`)},
	}

	for _, tc := range cases {
		data, err := json.Marshal(tc.value)
		if err != nil {
			t.Errorf("%s: json.Marshal() ошибка: %v", tc.name, err)
			continue
		}
		if !reflect.DeepEqual(data, tc.want) {
			t.Errorf("%s: json.Marshal() = %s, ожидалось %s", tc.name, data, tc.want)
		}
	}
}

// TestValueReceiverTextAndBinary Сериализация значений в текстовый и двоичный форматы
// обязана работать без взятия адреса. При смене приёмников MarshalText() или
// MarshalBinary() на *T значение перестанет реализовать соответствующие интерфейсы,
// и эти проверки не скомпилируются.
func TestValueReceiverTextAndBinary(t *testing.T) {
	var (
		text, _   = NewStringValue(stringTestBody).MarshalText()
		binary, _ = NewTimeValue(timeOkValidValue).MarshalBinary()
	)

	if string(text) != stringTestBody {
		t.Errorf("String.MarshalText() = %q, ожидалось %q", text, stringTestBody)
	}
	if len(binary) == 0 {
		t.Error("Time.MarshalBinary() вернул пустые данные")
	}
}
