package eventbus

import (
	"log"

	domainuser "starter/internal/domain/user"

	evbus "github.com/asaskevich/eventbus"
)

// Bus is a thin wrapper around asaskevich/eventbus that provides
// typed subscription and publishing for domain events.
type Bus struct {
	eb evbus.Bus
}

func New() *Bus {
	return &Bus{eb: evbus.New()}
}

func (b *Bus) Subscribe(topic string, fn interface{}) error {
	return b.eb.Subscribe(topic, fn)
}

func (b *Bus) SubscribeAsync(topic string, fn interface{}, transactional bool) error {
	return b.eb.SubscribeAsync(topic, fn, transactional)
}

func (b *Bus) WaitAsync() {
	b.eb.WaitAsync()
}

func (b *Bus) Publish(topic string, payload interface{}) {
	b.eb.Publish(topic, payload)
}

// PublishAll dispatches a slice of domain events.
func PublishAll[E interface{ EventName() string }](b *Bus, events []E) {
	for _, e := range events {
		b.Publish(e.EventName(), e)
	}
}

// RegisterDefaultHandlers wires up built-in logging handlers.
func (b *Bus) RegisterDefaultHandlers() {
	must(b.Subscribe("user.created", func(e domainuser.UserCreated) {
		log.Printf("[EVENT] %s – user=%s name=%q", e.EventName(), e.UserID, e.Name)
	}))
}

func must(err error) {
	if err != nil {
		panic("eventbus: failed to register handler: " + err.Error())
	}
}
