package events

// OnUserRegistered registers a listener for the "auth.user.registered" event.
func (b *Bus) OnUserRegistered(fn func(UserRegisteredPayload)) {
	b.on("auth.user.registered", func(v any) { fn(v.(UserRegisteredPayload)) })
}

// EmitUserRegistered publishes a UserRegisteredPayload.
func (b *Bus) EmitUserRegistered(p UserRegisteredPayload) {
	b.emit("auth.user.registered", p)
}