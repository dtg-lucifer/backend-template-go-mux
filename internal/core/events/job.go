package events

// OnJobEnqueued registers a listener for the "queue.job.enqueued" event.
func (b *Bus) OnJobEnqueued(fn func(JobEnqueuedPayload)) {
	b.on("queue.job.enqueued", func(v any) { fn(v.(JobEnqueuedPayload)) })
}

// EmitJobEnqueued publishes a JobEnqueuedPayload.
func (b *Bus) EmitJobEnqueued(p JobEnqueuedPayload) {
	b.emit("queue.job.enqueued", p)
}