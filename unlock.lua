--[[两个动作
 1. 检测是不是预期的值
 如果是删除，卜筮返回一个值
 ]]

if redis.Call("get",KEYS[1]) == ARGV[1] then
    return redis.Call("del",KEYS[1])
else
    -- key不是你的key不存在
    return 0
end

