package models

import "time"

// Program is a customer engagement that groups projects under one timeline.
type Program struct {
	BaseModel
	Name       string    `gorm:"column:name;not null"`
	Client     string    `gorm:"column:client;not null"`
	Status     string    `gorm:"column:status;not null;index"`
	StartDate  time.Time `gorm:"column:start_date;type:date;not null"`
	EndDate    time.Time `gorm:"column:end_date;type:date;not null"`
	Goal       string    `gorm:"column:goal;not null;default:''"`
	Objectives []string  `gorm:"column:objectives;serializer:json;type:jsonb;not null;default:'[]'"`
	CharterURL *string   `gorm:"column:charter_url"`
	LeadID     *string   `gorm:"column:lead_id;type:uuid;index"`

	Lead *Person `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
}
