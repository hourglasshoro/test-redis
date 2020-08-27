package main

func GetTotalNum() (res int64, err error) {
	redisInst, err := NewRedis()
	if err != nil {
		return
	}
	res, err = redisInst.DBSize(ctx).Result()
	return
}
