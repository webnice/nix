package nix

import "database/sql/driver"

// Interface Основный интерфейс, которому удовлетворяют все типы данных пакета.
// Интерфейсу удовлетворяют только указатели на типы: методы изменения и разбора
// объявлены на указателях, методы сериализации — на значениях.
type Interface interface {
	Reset()
	Scan(value any) error
	Value() (driver.Value, error)
	UnmarshalJSON(data []byte) error
	MarshalJSON() ([]byte, error)
	UnmarshalText(text []byte) error
	MarshalText() ([]byte, error)
	UnmarshalBinary(data []byte) error
	MarshalBinary() ([]byte, error)
}

// Valuer Интерфейс типизированного доступа к значению типа T.
// Дополняет Interface методами с параметром и возвратом типа T.
// Метод NullIfDefault() описывается отдельным интерфейсом Nullifier:
// он возвращает обёртку (nix.Time, а не time.Time), что невозможно выразить
// через параметр T.
type Valuer[T any] interface {
	Interface
	Set(value T)
	SetValid(value T)
	MustValue() T
	Pointer() *T
}

// Nullifier Интерфейс сброса значения по умолчанию до nil.
// W — тип обёртки: метод возвращает копию объекта после проверки значения.
// Для доступа к полному набору методов типа используется пересечение:
//
//	interface {
//		Valuer[T]
//		Nullifier[W]
//	}
type Nullifier[W any] interface {
	NullIfDefault() W
}
