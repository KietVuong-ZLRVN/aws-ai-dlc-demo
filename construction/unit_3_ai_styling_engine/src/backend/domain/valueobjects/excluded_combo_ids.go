package valueobjects

type ExcludedComboIds struct {
	ids map[string]bool
}

func NewExcludedComboIds(ids []string) ExcludedComboIds {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id != "" {
			m[id] = true
		}
	}
	return ExcludedComboIds{ids: m}
}

func (e ExcludedComboIds) Contains(id string) bool {
	return e.ids[id]
}

func (e ExcludedComboIds) Add(id string) ExcludedComboIds {
	newMap := make(map[string]bool, len(e.ids)+1)
	for k, v := range e.ids {
		newMap[k] = v
	}
	newMap[id] = true
	return ExcludedComboIds{ids: newMap}
}
