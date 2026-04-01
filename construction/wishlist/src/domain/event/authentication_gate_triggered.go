package event

import "time"

type AuthenticationGateTriggered struct {
	SimpleSku  string
	ReturnPath string
	OccurredAt time.Time
}

func (e AuthenticationGateTriggered) EventName() string {
	return "AuthenticationGateTriggered"
}
