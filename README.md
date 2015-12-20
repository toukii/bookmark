#	LFU BOOKMARK

Least frequently use bookmarks

[BOOKMARK](http://bookmark.daoapp.io/)

##	Bookmarks

[LFU](https://github.com/shaalx/leetcode)

[GoOJ](http://gooj-shaalx.myalauda.cn/)

##	Bug

这里存在一个这样的bug： 数据不可靠(solved @2015-11)

手动更新书签的时候，由于仅提供了Get,Set方法（均更新了书签的位置）,书签排名需要减去2；

然而，自动加2的时候，排名已经确定，仅通过排名减2，无法做到精确排名。

解决方法只能通过增加Attach,Put方法（透明更新书签）。

*	solution

	[cache.WhistPut(key, val)](https://github.com/shaalx/leetcode/blob/master/lfu2/lfu.go#L151)

##	LICENSE

Apache license 2.0