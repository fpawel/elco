package journal

import (
	"time"
)

//go:generate reform

// Entry represents a row in entry table.
//reform:entry
type Entry struct {
	EntryID   int64     `reform:"entry_id,pk"`
	WorkID    int64     `reform:"work_id"`
	CreatedAt time.Time `reform:"created_at"`
	Level     string    `reform:"level"`
	Message   string    `reform:"message"`
	File      string    `reform:"file"`
	Line      int64     `reform:"line"`
	Stack     string    `reform:"stack"`
}

// Entry represents a row in entry_info table.
//reform:entry_info
type EntryInfo struct {
	EntryID   int64     `reform:"entry_id,pk"`
	WorkID    int64     `reform:"work_id"`
	CreatedAt time.Time `reform:"created_at"`
	Level     string    `reform:"level"`
	Message   string    `reform:"message"`
	File      string    `reform:"file"`
	Line      int64     `reform:"line"`
	Stack     string    `reform:"stack"`
	WorkName  string    `reform:"work_name"`
}
