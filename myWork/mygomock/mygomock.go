package mygomock

//一个等待被测试的接口
type Repository interface{
	Create(key string, value []byte) error
	Retrieve(key string) ([]byte, error)
	Update(key string, value []byte) error
	Delete(key string) error
}
