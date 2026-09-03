package payment

// PixProvider is the port through which the application layer requests a
// simulated Pix "copy and paste" code and an asynchronous confirmation.
//
// Simulate must return immediately with a pixCode and schedule onApproved
// to be invoked exactly once, later, with paymentID — real implementations
// do this via a background goroutine with a randomized delay; test doubles
// may call it synchronously.
type PixProvider interface {
	Simulate(paymentID string, amount Money, onApproved func(paymentID string)) (pixCode string, err error)
}
