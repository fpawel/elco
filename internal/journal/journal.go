package journal

func (s Entry) EntryInfo(workName string) EntryInfo {
	return EntryInfo{
		CreatedAt: s.CreatedAt,
		EntryID:   s.EntryID,
		WorkID:    s.WorkID,
		Message:   s.Message,
		Level:     s.Level,
		File:      s.File,
		Line:      s.Line,
		Stack:     s.Stack,
		WorkName:  workName,
	}
}
