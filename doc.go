/*

Пакет nix — набор типов данных-обёрток для Go, позволяющих хранить в полях
структур «пустые» (nil) значения вместе с обычными.

Каждый тип реализует интерфейсы sql.Scanner и driver.Valuer (совместимость
с database/sql, GORM, sqlx и другими ORM), интерфейсы кодеков encoding/json,
encoding.TextMarshaler/TextUnmarshaler и encoding/gob. Пакет является
продолжением пакетов nul и lin.v1/nl: все типы, имена методов и их семантика
сохранены, добавлены новые типы и generic-вариант.

# Поддерживаемые типы данных

	Тип пакета    Базовый тип        В базе данных
	------------- ------------------ ----------------------------------------
	Bool          bool               bool
	String        string             string
	Bytes         []byte             []byte
	StringSlice   []string           set, строка "a,b,c"
	Int           int                int64
	Int8          int8               int64
	Int16         int16              int64
	Int32         int32              int64
	Int64         int64              int64
	Uint          uint               десятичная строка
	Uint8         uint8              десятичная строка
	Uint16        uint16             десятичная строка
	Uint32        uint32             десятичная строка
	Uint64        uint64             десятичная строка
	Uintptr       uintptr            десятичная строка
	Float32       float32            float64
	Float64       float64            float64
	Complex64     complex64          строка "(1.5+2.5i)"
	Complex128    complex128         строка "(1.5+2.5i)"
	Time          time.Time          datetime
	Duration      time.Duration      int64 (наносекунды); чтение из TIME
	                                 "[-]ЧЧЧ:ММ:СС[.дробная]"
	RawMessage    json.RawMessage    []byte
	IP            net.IP             строка "192.168.0.1"
	BigInt        *big.Int           десятичная строка
	UUID          [16]byte           строка "xxxxxxxx-…"
	Nullable[T]   скалярный T        как у соответствующего типа

# Соглашения

Каждый объект состоит из значения базового типа и флага действительности
Valid: Valid равен true если значение не nil и не значение по умолчанию для
типа данных. Нулевой объект (Valid=false) сериализуется как null (JSON),
строка "null" (текстовый формат) и как недействительный wrapper (gob).

Методы изменения и разбора объявлены на указателях, методы сериализации —
на значениях: интерфейсам пакета удовлетворяют только указатели на типы,
например *nix.Time.

Общие методы всех типов:

  - New<T>() — создание нулевого объекта;
  - New<T>Value(v) — создание действительного объекта из значения;
  - New<T>PointerValue(p) — создание объекта из ссылки на значение; nil даёт
    нулевой объект;
  - Set(v) — присвоение значения как есть, Valid=true, без проверок;
  - SetValid(v) — присвоение значения с установкой Valid=true; если присвоено
    значение по умолчанию, Valid сбрасывается;
  - Reset() — сброс значения и установка Valid=false;
  - NullIfDefault() — сброс значения до nil если оно является значением
    по умолчанию;
  - MustValue() — значение в любом случае; нулевой объект даёт значение
    по умолчанию;
  - Pointer() — ссылка на значение или nil для нулевого объекта.

# Сериализация

Действительное значение сериализуется как значение базового типа,
недействительное — как nil:

  - JSON: действительное значение кодируется как значение базового типа,
    нулевой объект — литералом null; при разборе дополнительно поддерживаются
    объекты {"<T>":…,"Valid":…};
  - текстовый формат: действительное значение — строковое представление
    базового типа, нулевой объект — строка "null";
  - двоичный формат (gob): структура wrapper.<T>Wrapper с полями значения
    и флага Valid.

# Работа с базами данных

Scan читает значения стандартных типов database/sql (nil, int64, float64,
bool, []byte, string, time.Time); чтение NULL (nil) даёт нулевой объект
с Valid=false. Value возвращает представление значения, принимаемое
драйверами (см. таблицу выше).

## Тип Duration и формат time базы данных

Тип Duration читается из колонок типа time баз данных MySQL и PostgreSQL,
которые драйверы передают строкой "[-][Ч]ЧЧ:ММ:СС[.дробная часть]":

  - MySQL TIME — длительность в диапазоне -838:59:59..838:59:59, в том числе
    с дробной частью (TIME(6), микросекунды); отрицательные значения
    передаются со знаком "-";
  - PostgreSQL time — время суток 00:00:00..24:00:00 с дробной частью
    до микросекунд.

Дробная часть принимается от 1 до 9 цифр и масштабируется до наносекунд.
Наряду с БД-форматом строки разбираются форматом time.ParseDuration
("1h30m0s", "300ms"), число — как наносекунды. БД-формат поддерживается
методами Scan, UnmarshalJSON и UnmarshalText. Запись в базу данных выполняется
методом Value как int64 наносекунд.

## Тип StringSlice и формат set базы данных

Тип StringSlice читается из колонок типа set баз данных MySQL, которые
драйверы передают строкой с перечислением элементов через запятую:

  - запись в базу данных методом Value выполняется строкой "a,b,c";
  - пустой слайс записывается пустой строкой — пустой набор set;
  - чтение разбирает строку set по запятым; пустая строка — пустой набор;
  - запятая и обратная косая черта в элементах экранируются обратной косой
    чертой: элемент "a,b" хранится как "a\,b", элемент "a\b" — как "a\\b";
    при чтении экранирование восстанавливается.

JSON-представление значения — массив строк, текстовое — строка set.

# Интерфейсы пакета

Все типы пакета (строго указатели) удовлетворяют интерфейсам:

  - Interface — общее подмножество: Reset, Scan, Value и кодеки JSON, Text
    и Binary; позволяет работать с гетерогенными коллекциями значений;
  - Valuer[T] — типизированный доступ: Set(T), SetValid(T), MustValue() T,
    Pointer() *T;
  - Nullifier[W] — NullIfDefault() W; метод возвращает обёртку (nix.Time,
    а не time.Time), поэтому описан отдельно от Valuer[T].

Полный набор методов типа выражается пересечением Valuer[T] и Nullifier[W].

# Пример использования

	// Message История запросов и ответов телеграм бота.
	type Message struct {
		ID       uint64         `db:"id"       gorm:"column:id;primary_key"`
		CreateAt time.Time      `db:"createAt" gorm:"column:createAt"`
		DeleteAt nix.Time       `db:"deleteAt" gorm:"column:deleteAt"`
		DataRaw  nix.RawMessage `db:"dataRaw"  gorm:"column:dataRaw"`
		Timeout  nix.Duration   `db:"timeout"  gorm:"column:timeout"`
		TraceID  nix.UUID       `db:"traceId"  gorm:"column:traceId"`
	}

	msg := Message{
		DeleteAt: nix.NewTimeValue(time.Now()),
		DataRaw:  nix.NewRawMessageValue([]byte(`{"update_id":1}`)),
		Timeout:  nix.NewDurationValue(90 * time.Minute),
	}
	_, err := db.Exec(
		"INSERT INTO message(deleteAt, dataRaw, timeout) VALUES(?,?,?)",
		msg.DeleteAt, msg.DataRaw, msg.Timeout,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Чтение: NULL корректно попадает в объект с Valid=false.
	var sendAt nix.Time
	if err = db.QueryRow("SELECT sendAt FROM message WHERE id=?", 1).Scan(&sendAt); err != nil {
		log.Fatal(err)
	}
	if !sendAt.Valid {
		// Значение в базе — NULL.
	}

# Требования

Go 1.18 или новее (используются generics). Зависимости отсутствуют.

*/

package nix
