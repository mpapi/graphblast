package graphblast

import (
	"encoding/json"
	"sync"
)

// Messages have two parts: an Envelope, which summarizes their contents, and
// Contents (which can be generated dynamically, and so might produce an
// error).
type Message interface {
	Recipient(string) bool
	Envelope() string
	Contents() ([]byte, error)
}

// NewJSONMessage makes a new message from an arbitrary object, the contents of
// which is the JSON representation of the object, to be sent to all
// recipients.
func NewJSONMessage(envelope string, contents interface{}) Message {
	return NewJSONMessageTo([]string{}, envelope, contents)
}

// NewJSONMessage makes a new message from an arbitrary object, the contents of
// which is the JSON representation of the object, to be sent to specific
// recipients.
func NewJSONMessageTo(recipients []string, envelope string, contents interface{}) Message {
	bytes, err := json.Marshal(contents)

	recipientSet := make(map[string]bool, len(recipients))
	for _, name := range recipients {
		recipientSet[name] = true
	}
	return staticMessage{recipientSet, envelope, bytes, err}
}

// staticMessage is the simplest implementation of Message -- the body is
// computed ahead of time.
type staticMessage struct {
	recipients map[string]bool
	envelope   string
	contents   []byte
	err        error
}

func (m staticMessage) Recipient(name string) bool {
	return len(m.recipients) == 0 || m.recipients[name]
}

func (m staticMessage) Envelope() string {
	return m.envelope
}

func (m staticMessage) Contents() ([]byte, error) {
	return m.contents, m.err
}

// Processes that need to receive messages can subscribe to (or unsubscribe
// from) a Publisher.
type Publisher interface {
	Subscribe(string) <-chan Message
	Unsubscribe(string)
}

// Processes that need to send messages can do so using Subscribers.
type Subscribers interface {
	Send(Message)
}

// broadcastRequest asks a Broadcaster to manipulate its map of subscribers.
type broadcastRequest func(map[string]chan<- Message)

// subscribeRequest asks a Broadcaster to add a new subscriber.
func subscribeRequest(name string, channel chan<- Message) broadcastRequest {
	return func(listeners map[string]chan<- Message) {
		listeners[name] = channel
	}
}

// subscribeRequest asks a Broadcaster to remove a subscriber.
func unsubscribeRequest(name string) broadcastRequest {
	return func(listeners map[string]chan<- Message) {
		delete(listeners, name)
	}
}

// A Broadcaster is a Publisher and Subscribers: receivers register with it,
// and messages sent through it are dispatched to all subscribers.
type Broadcaster struct {
	messages  chan Message
	listeners map[string]chan<- Message
	*sync.Mutex
}

// NewBroadcaster creates a new, synchronous Broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		messages:  make(chan Message),
		listeners: make(map[string]chan<- Message),
		Mutex:     new(sync.Mutex)}
}

// Subscribe adds a new subscriber by name to a Broadcaster, and returns a
// channel on which the subscriber should listen for Messages.
func (b *Broadcaster) Subscribe(name string) <-chan Message {
	result := make(chan Message)
	b.Lock()
	defer b.Unlock()
	b.listeners[name] = result
	return result
}

// Unsubscribe removes the subscriber from the Broadcaster, closing its channel
// so that it can no longer receive Messages.
func (b *Broadcaster) Unsubscribe(name string) {
	b.Lock()
	defer b.Unlock()
	delete(b.listeners, name)
}

// Send passes the message to all the Broadcaster's subscribers.
func (b *Broadcaster) Send(message Message) {
	b.messages <- message
}

// DispatchForever handles requests for new subscribers and dispatches sent
// messages to all subscribers.
func (b *Broadcaster) DispatchForever() {
	for message := range b.messages {
		for name, listener := range b.listeners {
			if !message.Recipient(name) {
				continue
			}
			listener <- message
		}
	}
}
