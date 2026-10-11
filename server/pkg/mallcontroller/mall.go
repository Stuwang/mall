package mallcontroller

import (
	"context"

	apimall "helloworld/api/mall"
	"helloworld/common/config"
)

var _ apimall.MallHTTPServer = (*Controller)(nil)

type Controller struct {
	pointsPerYuan int64
	mall          config.MallConfig
}

func New(pointsPerYuan int64, mall config.MallConfig) *Controller {
	return &Controller{
		pointsPerYuan: pointsPerYuan,
		mall:          mall,
	}
}

func (c *Controller) GetMallInfo(ctx context.Context, req *apimall.GetMallInfoRequest) (*apimall.GetMallInfoReply, error) {
	_ = ctx
	_ = req
	return &apimall.GetMallInfoReply{
		PointsPerYuan: c.pointsPerYuan,
		Contact: &apimall.ContactInfo{
			Phone:   c.mall.Contact.Phone,
			Email:   c.mall.Contact.Email,
			Wechat:  c.mall.Contact.Wechat,
			Address: c.mall.Contact.Address,
		},
		HelpCenter: c.mall.HelpCenter,
		AboutUs:    c.mall.AboutUs,
	}, nil
}
