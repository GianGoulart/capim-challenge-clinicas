package payment

// PixProvider é o port através do qual a camada de aplicação solicita um
// código Pix "copia e cola" simulado e uma confirmação assíncrona.
//
// Simulate deve retornar imediatamente com um pixCode e agendar a chamada
// de onApproved exatamente uma vez, mais tarde, com paymentID — implementações
// reais fazem isso via uma goroutine em background com um delay aleatório;
// test doubles podem chamá-lo de forma síncrona.
type PixProvider interface {
	Simulate(paymentID string, amount Money, onApproved func(paymentID string)) (pixCode string, err error)
}
