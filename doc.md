# 文档
## 为什么一定要设置过期时间 ?
分布式环境下，假如实例1持有一把锁，如果实例1崩溃，则其他等待该锁的实例会一直等待
不同于服务器关闭的优雅退出还有机会释放锁。
## 为什么使用uuid作为值
就是区分该锁是某个实例加的锁。当然其他的能够区分的值也是可以的

## 关于测试
严格来讲，单元测试不能依赖于任何第三方组件，也即不依赖真实的redis（只有集成测试才用）
单元测试，我们使用gomock工具
一个伏笔，代码使用了redis.Cmdable的接口，本质上是为了便于我们mock
