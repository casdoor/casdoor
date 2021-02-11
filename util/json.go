package util

import "encoding/json"

func StructToJson(v interface{}) string {
	//data, err := json.MarshalIndent(v, "", "  ")
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(data)
}
