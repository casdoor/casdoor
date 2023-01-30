// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

func Arrcmp(src []string, dest []string) ([]string, []string) {
	msrc := make(map[string]byte) //Indexing by source number
	mall := make(map[string]byte) //Indexing of all elements of the source + destination

	var set []string //Intersections

	//1.Source array build map
	for _, v := range src {
		msrc[v] = 0
		mall[v] = 0
	}
	//2.The set of all the elements that are not stored in the array of items, i.e., duplicate elements, is the merged set
	for _, v := range dest {
		l := len(mall)
		mall[v] = 1
		if l != len(mall) { //Length variation, i.e., can be stored
			l = len(mall)
		} else { //Can't save, enter merge
			set = append(set, v)
		}
	}
	//3.Iterate through the intersection, find it in the parallel set, delete it from the parallel set, and after the deletion, it is the complementary set (i.e., parallel-intersection = all changed elements)
	for _, v := range set {
		delete(mall, v)
	}
	//4.At this point, mall is the complementary set, all elements to the source to find, to find is to delete, can not find must be found in the destination array, that is, the newly added
	var added, deleted []string
	for v, _ := range mall {
		_, exist := msrc[v]
		if exist {
			deleted = append(deleted, v)
		} else {
			added = append(added, v)
		}
	}

	return added, deleted
}
