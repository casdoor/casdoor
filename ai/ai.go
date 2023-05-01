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
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

func queryAnswer(authToken string, question string, timeout int) (string, error) {
	//fmt.Printf("Question: %s\n", question)

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
		//fmt.Printf("%s\n", err.Error())
		return "", err
	}

	res := resp.Choices[0].Message.Content
	res = strings.Trim(res, "\n")
	//fmt.Printf("Answer: %s\n\n", res)
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
