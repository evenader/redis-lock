package redis_lock

import (
	"context"
	_ "embed"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

// 基于redis的分布式锁 排他式的key/value

var (
	ErrLockNotHold = errors.New("未持有锁")

	ErrFailedToPreemptLock = errors.New("加锁失败")
)

// import  对用户暴露的操作地方
type Client struct {
	client redis.Cmdable
}

func NewClient(r redis.Cmdable) *Client {
	//		正常应该返回error 但是这里没放 因为显然这里不需要
	return &Client{
		client: r,
	}
}

var (
	//go embed
	luaUnlock string
)

type Lock struct {
	client redis.Cmdable
	key    string
	value  string
}

func newLock(client redis.Cmdable, key string) *Lock {
	return &Lock{
		client: client,
		key:    key,
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	// uuid 是唯一的
	value := uuid.New().String()

	res, err := c.client.SetNX(ctx, key, value, time.Minute).Result()
	if err != nil {
		return nil, err
	}
	if !res {
		return nil, ErrFailedToPreemptLock
	}
	return newLock(c.client, key), nil
}

func (l *Lock) UnLock(ctx context.Context, key string, expiration time.Duration) error {
	/*
				存在一个问题 就是假定一种情况
				t1时刻 val还是你设置的 你拿过来
				t2时刻 有人篡改了cal
				t3时刻 你删除了key
		val, err := l.client.Get(ctx, key).Result()
			if err != nil {
				return err
			}
			if l.value != val {
					_, err := l.client.Del(ctx, key).Result()
					if err != nil {
						return err
					}
				}
				return nil
	*/
	/*
		因此考虑引入lua脚本
	*/
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()
	if err == redis.Nil {
		return ErrLockNotHold
	}
	if err != nil {
		return err
	}
	if res == 0 {
		//key不存在 锁不是你的
		return ErrLockNotHold
	}

	return nil
}

func (c *Client) Lock_Wrong(ctx context.Context, key string) error {
	// todo 这里设置了一个过期时间 但是原因留白
	res, err := c.client.SetNX(ctx, key, "123", time.Minute).Result()
	if err != nil {
		return nil
	}
	if !res {
		return errors.New("加锁失败")
	}
	return nil
}

func (c *Client) UnLock_Wrong(ctx context.Context, key string) error {
	// todo 这里是按照反例举的 这个写法不对
	/*
		为什么删除的时候不能这样
		因为 假设一种场景 a成功申请到一把锁 设置过期时间为10s 但是10s后他还在用
		此时 因为过期 b也申请到这把锁
		那么a后续删除时必须确保删除的锁还是他自己的
	*/
	res, err := c.client.Del(ctx, key).Result()
	if err != nil {
		return nil
	}
	if res != 1 {
		// 过期了 / key被人删除了
		return errors.New("加锁失败")
	}
	return nil
}
