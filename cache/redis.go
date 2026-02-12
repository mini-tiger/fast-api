package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/mini-tiger/fast-api/config"
	"github.com/mini-tiger/fast-api/dError"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var ctx = context.Background()

func init() {
	redisConfig := config.GetInstance().Section("redis")
	host := redisConfig.Key("host").Value()
	port := redisConfig.Key("port").Value()
	password := redisConfig.Key("password").Value()
	if len(password) == 0 {
		password = redisConfig.Key("password1").Value() + "#" + redisConfig.Key("password2").Value()
	}
	db := redisConfig.Key("db").MustInt(0)

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(dError.NewError("连接Redis出错", err))
	}
}

// SetWithExpire 设置键值对，并指定过期时间
// 支持字符串、数字、布尔值、切片、结构体、map等类型
// 对于复杂类型（切片、结构体、map），会自动使用JSON序列化
func SetWithExpire(key string, value interface{}, expiration time.Duration) error {
	data, err := serializeValue(value)
	if err != nil {
		return fmt.Errorf("序列化值失败: %v", err)
	}
	return redisClient.Set(ctx, key, data, expiration).Err()
}

// serializeValue 序列化值，对于复杂类型使用JSON，简单类型直接转换
func serializeValue(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	// 获取值的类型
	v := reflect.ValueOf(value)
	kind := v.Kind()

	// 处理指针类型
	if kind == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
		kind = v.Kind()
	}

	// 简单类型直接转换为字符串
	switch kind {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float()), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool()), nil
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct, reflect.Interface:
		// 复杂类型使用JSON序列化
		jsonData, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("JSON序列化失败: %v", err)
		}
		// 添加特殊前缀标识这是JSON数据
		return "JSON:" + string(jsonData), nil
	default:
		// 其他类型尝试JSON序列化
		jsonData, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("不支持的类型 %v: %v", kind, err)
		}
		return "JSON:" + string(jsonData), nil
	}
}

// Get 获取字符串值（向后兼容，返回原始字符串）
func Get(key string) (string, error) {
	result, err := redisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("键 %s 不存在", key)
	}
	return result, err
}

// GetObject 获取值并反序列化到指定类型
// 支持字符串、数字、布尔值、切片、结构体、map等类型
// 示例: var user User; err := GetObject("user:1", &user)
func GetObject(key string, dest interface{}) error {
	data, err := redisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("键 %s 不存在", key)
	}
	if err != nil {
		return err
	}

	return deserializeValue(data, dest)
}

// deserializeValue 反序列化值
func deserializeValue(data string, dest interface{}) error {
	if dest == nil {
		return fmt.Errorf("目标对象不能为nil")
	}

	// 检查是否是JSON格式的数据
	if len(data) >= 5 && data[:5] == "JSON:" {
		// JSON格式，直接反序列化
		jsonData := data[5:]
		return json.Unmarshal([]byte(jsonData), dest)
	}

	// 简单类型，需要根据目标类型进行转换
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("目标对象必须是指针类型")
	}

	destElem := destValue.Elem()
	destKind := destElem.Kind()

	switch destKind {
	case reflect.String:
		destElem.SetString(data)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 根据目标类型确定位数
		var bitSize int
		switch destKind {
		case reflect.Int8:
			bitSize = 8
		case reflect.Int16:
			bitSize = 16
		case reflect.Int32:
			bitSize = 32
		case reflect.Int, reflect.Int64:
			bitSize = 64
		}
		intVal, err := strconv.ParseInt(data, 10, bitSize)
		if err != nil {
			return fmt.Errorf("无法将 %s 转换为整数: %v", data, err)
		}
		destElem.SetInt(intVal)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			return fmt.Errorf("无法将 %s 转换为无符号整数: %v", data, err)
		}
		destElem.SetUint(uintVal)
		return nil
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(data, 64)
		if err != nil {
			return fmt.Errorf("无法将 %s 转换为浮点数: %v", data, err)
		}
		destElem.SetFloat(floatVal)
		return nil
	case reflect.Bool:
		// 支持多种布尔值格式
		var boolVal bool
		switch data {
		case "true", "1", "True", "TRUE", "yes", "Yes", "YES":
			boolVal = true
		case "false", "0", "False", "FALSE", "no", "No", "NO":
			boolVal = false
		default:
			// 尝试解析为布尔值
			parsed, err := strconv.ParseBool(data)
			if err != nil {
				return fmt.Errorf("无法将 %s 转换为布尔值: %v", data, err)
			}
			boolVal = parsed
		}
		destElem.SetBool(boolVal)
		return nil
	default:
		// 其他类型尝试JSON反序列化
		return json.Unmarshal([]byte(data), dest)
	}
}

// GetBytes 获取字节数组值
func GetBytes(key string) ([]byte, error) {
	result, err := redisClient.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("键 %s 不存在", key)
	}
	return result, err
}

// Delete 删除一个或多个键
func Delete(keys ...string) error {
	return redisClient.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	count, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Expire 设置键的过期时间
func Expire(key string, expiration time.Duration) error {
	return redisClient.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余过期时间（秒）
func TTL(key string) (time.Duration, error) {
	return redisClient.TTL(ctx, key).Result()
}

// Increment 将键的值增加1
func Increment(key string) (int64, error) {
	return redisClient.Incr(ctx, key).Result()
}

// IncrementBy 将键的值增加指定数值
func IncrementBy(key string, value int64) (int64, error) {
	return redisClient.IncrBy(ctx, key, value).Result()
}

// Decrement 将键的值减少1
func Decrement(key string) (int64, error) {
	return redisClient.Decr(ctx, key).Result()
}

// DecrementBy 将键的值减少指定数值
func DecrementBy(key string, value int64) (int64, error) {
	return redisClient.DecrBy(ctx, key, value).Result()
}

// HSet 设置哈希字段值
func HSet(key string, field string, value interface{}) error {
	return redisClient.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希字段值
func HGet(key string, field string) (string, error) {
	result, err := redisClient.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("哈希字段 %s.%s 不存在", key, field)
	}
	return result, err
}

// HGetAll 获取哈希的所有字段和值
func HGetAll(key string) (map[string]string, error) {
	return redisClient.HGetAll(ctx, key).Result()
}

// HDel 删除哈希的一个或多个字段
func HDel(key string, fields ...string) error {
	return redisClient.HDel(ctx, key, fields...).Err()
}

// LPush 从列表左侧推入元素
func LPush(key string, values ...interface{}) error {
	return redisClient.LPush(ctx, key, values...).Err()
}

// RPush 从列表右侧推入元素
func RPush(key string, values ...interface{}) error {
	return redisClient.RPush(ctx, key, values...).Err()
}

// LPop 从列表左侧弹出元素
func LPop(key string) (string, error) {
	result, err := redisClient.LPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("列表 %s 为空", key)
	}
	return result, err
}

// RPop 从列表右侧弹出元素
func RPop(key string) (string, error) {
	result, err := redisClient.RPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("列表 %s 为空", key)
	}
	return result, err
}

// LRange 获取列表指定范围内的元素
func LRange(key string, start, stop int64) ([]string, error) {
	return redisClient.LRange(ctx, key, start, stop).Result()
}

// SAdd 向集合添加成员
func SAdd(key string, members ...interface{}) error {
	return redisClient.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合的所有成员
func SMembers(key string) ([]string, error) {
	return redisClient.SMembers(ctx, key).Result()
}

// SIsMember 检查成员是否在集合中
func SIsMember(key string, member interface{}) (bool, error) {
	return redisClient.SIsMember(ctx, key, member).Result()
}

// SRem 从集合中移除成员
func SRem(key string, members ...interface{}) error {
	return redisClient.SRem(ctx, key, members...).Err()
}

// Keys 根据模式查找所有匹配的键
func Keys(pattern string) ([]string, error) {
	return redisClient.Keys(ctx, pattern).Result()
}

// FlushDB 清空当前数据库
func FlushDB() error {
	return redisClient.FlushDB(ctx).Err()
}

// Close 关闭Redis连接
func Close() error {
	return redisClient.Close()
}

// Lock 非阻塞锁，获取成功返回 true
func Lock(key string, value any, expiration time.Duration) (bool, error) {
	result, err := redisClient.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}
func Unlock(key string) {
	_ = redisClient.Del(ctx, key).Err()
}

// BlockingLock 阻塞锁：在 waitTimeout 内轮询获取锁，获取成功返回 true；超时返回 false, nil
func BlockingLock(key string, lockExpiration time.Duration) (bool, error) {
	return BlockingLockWithInterval(key, 1, lockExpiration, 60*time.Second, 50*time.Millisecond)
}

// BlockingLockWithInterval
// key 锁键；value 锁的值（建议唯一，解锁时需传入同一 value 调用 Unlock）；
// lockExpiration 锁过期时间；waitTimeout 最长等待时间
// 轮询间隔为 retryInterval，默认 50ms
func BlockingLockWithInterval(key string, value interface{}, lockExpiration, waitTimeout, retryInterval time.Duration) (bool, error) {
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		ok, err := Lock(key, value, lockExpiration)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		time.Sleep(retryInterval)
	}
	return false, nil
}
