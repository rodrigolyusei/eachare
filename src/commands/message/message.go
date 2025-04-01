package message

// Estrutura para armazenar as informações da mensagem
type BaseMessage struct {
	Origin    string
	Clock     int
	Type      MessageType
	Arguments []string
}
