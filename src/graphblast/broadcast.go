package graphblast

// EventSource requests that want to listen for errors (including EOF) can
// register themselves here.
type ErrorWatchers map[string]chan error

// Adds a watcher by name to the map, and returns a new channel they can use
// for listening.
func (ew *ErrorWatchers) Watch(name string) chan error {
	errChan := make(chan error)
	(*ew)[name] = errChan
	return errChan
}

// Removes the watcher by name from the map.
func (ew *ErrorWatchers) Unwatch(name string) {
	errChan, ok := (*ew)[name]
	if ok {
		close(errChan)
		delete(*ew, name)
	}
}

// Takes a channel of errors, and broadcasts any errors that are received on
// that channel to all registered channels.
func (ew *ErrorWatchers) Broadcast(errors chan error) {
	for err := range errors {
		for _, errChan := range *ew {
			errChan <- err
		}
	}
}
