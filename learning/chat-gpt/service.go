package chatgpt

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

type service struct {
	chatGptClient *openai.Client
}

func NewService(t string) *service {
	client := openai.NewClient(
		option.WithAPIKey(t), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	return &service{
		chatGptClient: &client,
	}
}

func (s *service) CodeReview(code string) (string, error) {
	// Implement the logic to interact with the ChatGPT API and get a response
	// This is a placeholder implementation

	codeReviewRoleAndPersona := `
	You are an elite Google-level software engineer. 
	Your job is to provide precise, high-quality code reviews.
	Your input is a git diff. Output your review in json format as below (list of comments by line):
	{
		comments: [
			{
				body: your review,
				line: your comment in a new version,
				path: the file path,
				side: RIGHT // only comment on the new version and should always be RIGHT
			}
		]
	}.
	When commenting, Focus on the new code lines and ignore the old code lines.
	`

	chatCompletion, err := s.chatGptClient.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(codeReviewRoleAndPersona),
			openai.UserMessage(code),
		},
		Model:       "gpt-4",
		Temperature: param.Opt[float64]{Value: 0.3},
		MaxTokens:   param.Opt[int64]{Value: 2048},
		TopP:        param.Opt[float64]{Value: 0.2},
	})

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	review := chatCompletion.Choices[0].Message.Content
	fmt.Println(review)
	return review, nil
}
