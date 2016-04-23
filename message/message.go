package message // import "engo.io/engo/message"

var (
	Mailbox *MessageManager
)

type MessageHandler func(msg Message)

type Message interface {
	Type() string
}

type MessageManager struct {
	// TODO(u): Unexport Listeners.
	Listeners map[string][]MessageHandler
}

func (mm *MessageManager) Dispatch(message Message) {
	handlers := mm.Listeners[message.Type()]

	for _, handler := range handlers {
		handler(message)
	}
}

func (mm *MessageManager) Listen(messageType string, handler MessageHandler) {
	if mm.Listeners == nil {
		mm.Listeners = make(map[string][]MessageHandler)
	}
	mm.Listeners[messageType] = append(mm.Listeners[messageType], handler)
}
