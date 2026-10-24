// Package biz 放发送、更新、历史的事务。当前只有占位，分配序号和预览开关尚未实现。
package biz

import "mall/im/pkg/bizerror"

func NotImplemented() error {
	return bizerror.ErrorNotImplemented
}
