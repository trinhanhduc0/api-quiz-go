package initialize

func Run() {

	LoadConfig()  //Load .env
	InitMongoDB() //Init Mongo db
	InitRedis()   //Init redis
	InitRouter()  //Init Router

}
