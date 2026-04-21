# Plan to Issue

当你看到此Rule的时候你应该正处于刚结束`brainstorming`（Spec文件已经Fixed）准备进入`writing-plans`的阶段。
在进入`writing-plans`之前你需要先检查`docs/superpowers/specs/yyyy-mm-dd-title.md`的顶部是否有一段HTML Comment记录了此Spec文件链接的issue的URL。
如果没有找到相关内容，则使用`gh-cli`命令将刚才完成的`docs/superpowers/specs/yyyy-mm-dd-title.md`作为一个issue提交到Github，具体例子如下：
假设

```docs/superpowers/specs/yyyy-mm-dd-title.md
# 添加黑暗模式

foobar...
```

的话，issue的标题就设为`添加黑暗模式`，issue正文则是`foobar...`。
完成issue提交后，在`docs/superpowers/specs/yyyy-mm-dd-title.md`的顶部写一段HTML Comment记录此文件链接的issue的URL。
