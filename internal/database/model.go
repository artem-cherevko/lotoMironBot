package database

import "github.com/lib/pq"

// type User struct {
// 	ID int64 `gorm:"primaryKey;not null"`

// }

type Collections string

const (
	Standard  Collections = "standard"
	NewYear   Collections = "new-year"
	Halloween Collections = "halloween"
)

type Ticket struct {
	ID     uint   `gorm:"primaryKey"`
	FileID string `gorm:"not null"`

	Numbers    pq.StringArray `gorm:"type:text[];not null"`
	Collection Collections    `gorm:"type:varchar(32);not null;default:'standard'"`
}
