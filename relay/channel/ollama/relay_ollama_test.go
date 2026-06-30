package ollama

import (
	"encoding/base64"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	"github.com/stretchr/testify/require"
)

func TestOpenAIChatToOllamaChat_RejectsUnsupportedFileAttachment(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{
		Model: "llama3.2-vision",
		Messages: []dto.Message{
			{
				Role: "user",
				Content: []any{
					map[string]any{
						"type": "text",
						"text": "Review this text file",
					},
					map[string]any{
						"type": "file",
						"file": map[string]any{
							"filename":  "note.txt",
							"file_data": "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte("hello from attachment")),
						},
					},
				},
			},
		},
	}

	_, err := openAIChatToOllamaChat(nil, request)
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported non-image attachment for Ollama")
}
