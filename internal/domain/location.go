package domain

type Location int

const (
	LocationUnknown  Location = iota // 0 = "не задано" - ловушка для багов
	LocationProducts                 // "Продукты"
	LocationPickup                   // ПВЗ марктеплейсов
)
