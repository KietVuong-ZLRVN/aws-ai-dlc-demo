package events

import "ai-styling-engine/domain/valueobjects"

const EventTypeCatalogItemsFetched = "CatalogItemsFetched"

type CatalogItemsFetched struct {
	SessionId    valueobjects.StyleSessionId
	CatalogItems []valueobjects.ComboItem
}

func (e CatalogItemsFetched) EventType() string {
	return EventTypeCatalogItemsFetched
}
