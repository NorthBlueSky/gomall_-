package dal

import (
	"github.com/qitian118/gomall/app/email/biz/dal/mysql"
	"github.com/qitian118/gomall/app/email/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
