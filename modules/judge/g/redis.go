// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package g

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var RedisConnPool *redis.Pool

func InitRedisConnPool() {
	if !Config().Alarm.Enabled {
		return
	}

	auth, dsn, db := formatRedisAddr(Config().Alarm.Redis.Dsn)
	maxIdle := Config().Alarm.Redis.MaxIdle
	idleTimeout := 240 * time.Second

	connTimeout := time.Duration(Config().Alarm.Redis.ConnTimeout) * time.Millisecond
	readTimeout := time.Duration(Config().Alarm.Redis.ReadTimeout) * time.Millisecond
	writeTimeout := time.Duration(Config().Alarm.Redis.WriteTimeout) * time.Millisecond

	RedisConnPool = &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			var opts []redis.DialOption
			if auth != "" {
				opts = append(opts, redis.DialPassword(auth))
			}
			if db != "" {
				if dbValue, err := strconv.ParseInt(db, 10, 32); err == nil {
					opts = append(opts, redis.DialDatabase(int(dbValue)))
				}
			}
			opts = append(opts, redis.DialConnectTimeout(connTimeout))
			opts = append(opts, redis.DialReadTimeout(readTimeout))
			opts = append(opts, redis.DialWriteTimeout(writeTimeout))
			c, err := redis.Dial("tcp", dsn, opts...)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: PingRedis,
	}
}

func formatRedisAddr(addrConfig string) (string, string, string) {
	redisAddr := strings.Split(addrConfig, "@")
	auth := ""
	host := redisAddr[len(redisAddr)-1]
	db := ""
	if len(redisAddr) > 1 {
		auth = strings.Join(redisAddr[0:len(redisAddr)-1], "@")
	}
	redisAddr = strings.Split(host, "/")
	if len(redisAddr) > 1 {
		host = redisAddr[0]
		db = redisAddr[1]
	}
	return auth, host, db
}

func PingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	if err != nil {
		log.Println("[ERROR] ping redis fail", err)
	}
	return err
}
