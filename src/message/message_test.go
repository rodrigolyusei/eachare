package message

import "testing"

func TestGetCommandType(t *testing.T) {
	tests := []struct {
		input    string
		expected MessageType
	}{
		{"GET_PEERS", GET_PEERS},
		{"PEERS_LIST", PEERS_LIST},
		{"UNKNOWN_COMMAND", UNKNOWN},
	}

	for _, test := range tests {
		result := GetMessageType(test.input)
		if result != test.expected {
			t.Errorf("GetCommandType(%s) = %d; expected %d", test.input, result, test.expected)
		}
	}
}

func TestCommandTypeString(t *testing.T) {
	tests := []struct {
		input    MessageType
		expected string
	}{
		{GET_PEERS, "GET_PEERS"},
		{PEERS_LIST, "PEERS_LIST"},
		{UNKNOWN, "UNKNOWN"},
	}

	for _, test := range tests {
		result := test.input.String()
		if result != test.expected {
			t.Errorf("CommandType(%d).String() = %s; expected %s", test.input, result, test.expected)
		}
	}
}
