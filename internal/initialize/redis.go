package initialize

import (
	models "my_app/internal/model"
)

func InitRedis() {
	models.GetRedis()
}
