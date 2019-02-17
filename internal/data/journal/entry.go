package journal

import (
	"time"
)

//go:generate reform

// Entry represents a row in entry table.
//reform:entry
type Entry struct {
	EntryID   int64     `reform:"entry_id,pk"`
	CreatedAt time.Time `reform:"created_at"`
	Message   string    `reform:"message"`
	Level     string    `reform:"level"`
	WorkID    int64     `reform:"work_id"`
}

// Entry represents a row in entry_info table.
//reform:entry_info
type EntryInfo struct {
	EntryID   int64     `reform:"entry_id,pk"`
	CreatedAt time.Time `reform:"created_at"`
	Message   string    `reform:"message"`
	Level     string    `reform:"level"`
	WorkID    int64     `reform:"work_id"`
	WorkName  string    `reform:"work_name"`
}
