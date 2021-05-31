package util

import (
	"github.com/astaxie/beego"
	"log"

	"github.com/garyburd/redigo/redis"
)

func SetValue(key string, value interface{}) {
	conn, err := redis.Dial(beego.AppConfig.String("redisNetwork"), beego.AppConfig.String("redisAddress"))

	if err != nil {
		log.Fatal(err.Error())
	}
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	_, e := conn.Do("SET", key, value)
	if e != nil {
		log.Fatal(e.Error())
	}
}

func GetValue(key string) interface{} {
	conn, err := redis.Dial(beego.AppConfig.String("redisNetwork"), beego.AppConfig.String("redisAddress"))

	if err != nil {
		log.Fatal(err.Error())
	}
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	value, e := conn.Do("GET", key)
	if e != nil {
		log.Fatal(e.Error())
	}
	return value
}
