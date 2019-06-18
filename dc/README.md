# dc - Deep Copy


### Usage:

```go
# 用于 gorm 复制
err := dc.GormDefaultCopier.Copy(to, from, "Id")

# 用于标准的深拷贝
err := dc.StandardCopier.Copy(to, from, "IgnoredField1", "IgnoredField2")

```

### GORM 复制

和标准的深拷贝不同之处在于：

1. from.field 为 nil 时，to.field 值保持不变
2. from.field == to.field 值时，to.field 被设置为 golang default value.

这样做是为了让目标结构 `to` 的内容仅保持和 `from` 不同的内容，从而可以得到一个变更记录，该记录被提交给 gorm 引擎时能够仅 Update 改变了的内容，而未变化的内容则不必再次被提交Update。

