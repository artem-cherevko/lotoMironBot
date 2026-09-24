package fsm

import (
	"lotoMironBot/internal/database"

	"github.com/go-telegram/fsm"
)

const (
	StateDefault fsm.StateID = "default"

	// Add tickets
	StateSelectCollection fsm.StateID = "select_collection"
	StateProvidePhotos    fsm.StateID = "provide_photos"
	StateFinish           fsm.StateID = "finish"

	DataFileIDs = "data-file-ids"
)

type AddTicketsData struct {
	Collection database.Collections
	FileIDs    []string
}
