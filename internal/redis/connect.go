package redis

import redis2 "github.com/redis/go-redis/v9"

func Connect(addr string) (*redis2.Client, error) {
	opt, err := redis2.ParseURL(addr)
	if err != nil {
		return nil, err
	}

	return redis2.NewClient(opt), nil
}
