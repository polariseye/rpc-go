# rpc-go
使用Golang实现的RPC

# 底层协议设计
{HEAD(1bytes)}{Flag(1Byte)}{RequestFrameId:(4Byte)}{ResponseFrameId:(4Byte)}{ContentLength:(4Byte)}{MethodNameLen:(1Byte)}{Tail:(1Byte)}{MethodName}{Content}

说明：
1. 如果是应答，可以不设置方法名
2. Flag:用于内容扩展字段 {数据包类型:2bit}{是否出错:1bit}{是否需要应答:1bit}{未使用:4bit}
# 接口设计
要求：
1. 能够使用基本接口简单包装出上层调用的接口
2. 能够支持异步调用
3. 能够传输流对象-->上层自己实现，协议和连接层不考虑这个问题
4. 能够对连接两边都实现这个（不区分客户端还是服务端）

# 还需要考虑的问题
* 断线重连
* 心跳处理 -->已添加
* server端的连接管理
* 需要实现一个自定义的序列化反序列化convertor
* 需要支持纯字节流的传输

# 断线重连需要考虑的问题
1. 发送方数据正确送达保障，Server和Client两边都需要保障
