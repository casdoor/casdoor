// Copyright 2023 The casbin Authors. All Rights Reserved.
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

package util

// Stack is a stack data structure implemented using a slice
type Stack struct {
	items []interface{}
}

// Push adds an item to the stack
func (s *Stack) Push(item interface{}) {
	s.items = append(s.items, item)
}

// Pop removes and returns the last item from the stack
func (s *Stack) Pop() (interface{}, bool) {
	if len(s.items) == 0 {
		return nil, false // Return a sentinel value or you could handle this more gracefully
	}
	lastIndex := len(s.items) - 1
	item := s.items[lastIndex]
	s.items = s.items[:lastIndex]
	return item, true
}

// Peek returns the last item from the stack without removing it
func (s *Stack) Peek() interface{} {
	if len(s.items) == 0 {
		return -1
	}
	return s.items[len(s.items)-1]
}

// IsEmpty checks if the stack is empty
func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of items in the stack
func (s *Stack) Size() int {
	return len(s.items)
}

// NewStack creates a new stack
func NewStack() *Stack {
	return &Stack{}
}
