package model

type SupportedLanguage struct {
	ID             uint   `gorm:"primaryKey"`
	Locale         string `gorm:"uniqueIndex;size:16;not null"`
	Name           string `gorm:"size:64;not null"`
	NativeName     string `gorm:"size:64;not null"`
	IsEnabled      int    `gorm:"default:1"`
	SortOrder      int    `gorm:"default:0"`
	IsDefault      int    `gorm:"default:0"`
	FallbackLocale string `gorm:"default:en;size:16"`
}

type ModelVariant struct {
	ID               uint    `gorm:"primaryKey"`
	BaseModel        string  `gorm:"size:64;not null"`
	Suffix           string  `gorm:"size:32;not null"`
	ProviderID       uint    `gorm:"index"`
	PriceCoefficient float64 `gorm:"default:1.0"`
	Priority         int     `gorm:"default:0"`
}
