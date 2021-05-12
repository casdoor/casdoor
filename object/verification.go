package object

import (
	"fmt"
	"math/rand"
	"time"
)

type VerificationRecord struct {
	RemoteAddr string `xorm:"varchar(100) notnull pk"`
	Receiver   string `xorm:"varchar(100) notnull"`
	Code       string `xorm:"varchar(10) notnull"`
	Time       int64  `xorm:"notnull"`
	IsUsed     bool
}

func SendVerificationCodeToEmail(remoteAddr, dest string) string {
	title := "Casdoor Code"
	sender := "Casdoor Admin"
	code := getRandomCode(5)
	content := fmt.Sprintf("You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes.", code)

	if result := AddToVerificationRecord(remoteAddr, dest, code); len(result) != 0 {
		return result
	}

	if err := SendEmail(title, content, dest, sender); err != nil {
		panic(err)
	}

	return ""
}

func AddToVerificationRecord(remoteAddr, dest, code string) string {
	var record VerificationRecord
	record.RemoteAddr = remoteAddr
	has, err := adapter.Engine.Get(&record)
	if err != nil {
		panic(err)
	}

	now := time.Now().Unix()

	if has && now - record.Time < 60 {
		return "You can only send one code in 60s."
	}

	record.Receiver = dest
	record.Code = code
	record.Time = now
	record.IsUsed = false

	if has {
		_, err = adapter.Engine.ID(record.RemoteAddr).AllCols().Update(record)
	} else {
		_, err = adapter.Engine.Insert(record)
	}

	if err != nil {
		panic(err)
	}

	return ""
}

func CheckVerificationCode(dest, code string) string {
	var record VerificationRecord
	record.Receiver = dest
	has, err := adapter.Engine.Desc("time").Where("is_used = 0").Get(&record)
	if err != nil {
		panic(err)
	}

	if !has {
		return "Code has not been sent yet!"
	}

	now := time.Now().Unix()
	if now-record.Time > 5*60 {
		return "You should verify your code in 5 min!"
	}

	if record.Code != code {
		return "Wrong code!"
	}

	record.IsUsed = true
	_, err = adapter.Engine.ID(record.RemoteAddr).AllCols().Update(record)
	if err != nil {
		panic(err)
	}

	return ""
}

// from Casnode/object/validateCode.go line 116
var stdNums = []byte("0123456789")

func getRandomCode(length int) string {
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, stdNums[r.Intn(len(stdNums))])
	}
	return string(result)
}
