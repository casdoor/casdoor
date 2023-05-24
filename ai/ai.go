// Copyright 2023 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

func queryAnswer(authToken string, question string, timeout int) (string, error) {
	// fmt.Printf("Question: %s\n", question)

	client := getProxyClientFromToken(authToken)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2+timeout*2)*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: question,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	res := resp.Choices[0].Message.Content
	res = strings.Trim(res, "\n")
	// fmt.Printf("Answer: %s\n\n", res)
	return res, nil
}

func QueryAnswerSafe(authToken string, question string) string {
	var res string
	var err error
	for i := 0; i < 10; i++ {
		res, err = queryAnswer(authToken, question, i)
		if err != nil {
			if i > 0 {
				fmt.Printf("\tFailed (%d): %s\n", i+1, err.Error())
			}
		} else {
			break
		}
	}
	if err != nil {
		panic(err)
	}

	return res
}

func QueryAnswerStream(authToken string, question string, writer io.Writer, builder *strings.Builder) error {
	client := getProxyClientFromToken(authToken)

	ctx := context.Background()
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("writer does not implement http.Flusher")
	}
	// https://platform.openai.com/tokenizer
	// https://github.com/pkoukk/tiktoken-go#available-encodings
	promptTokens, err := getTokenSize(openai.GPT3TextDavinci003, question)
	if err != nil {
		return err
	}

	// https://platform.openai.com/docs/models/gpt-3-5
	maxTokens := 4097 - promptTokens

	respStream, err := client.CreateCompletionStream(
		ctx,
		openai.CompletionRequest{
			Model:     openai.GPT3TextDavinci003,
			Prompt:    question,
			MaxTokens: maxTokens,
			Stream:    true,
		},
	)
	if err != nil {
		return err
	}
	defer respStream.Close()

	isLeadingReturn := true
	for {
		completion, streamErr := respStream.Recv()
		if streamErr != nil {
			if streamErr == io.EOF {
				break
			}
			return streamErr
		}

		data := completion.Choices[0].Text
		if isLeadingReturn && len(data) != 0 {
			if strings.Count(data, "\n") == len(data) {
				continue
			} else {
				isLeadingReturn = false
			}
		}

		fmt.Printf("%s", data)

		// Write the streamed data as Server-Sent Events
		if _, err = fmt.Fprintf(writer, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		// Append the response to the strings.Builder
		builder.WriteString(data)
	}

	return nil
}
