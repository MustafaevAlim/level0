package model

func OrderToResponse(o *Order) OrderInResponse {
	newOrder := OrderInResponse{
		OrderUID:          o.OrderUID,
		TrackNumber:       o.TrackNumber,
		Entry:             o.Entry,
		Locale:            o.Locale,
		InternalSignature: o.InternalSignature,
		CustomerId:        o.CustomerId,
		DeliveryService:   o.DeliveryService,
		CreatedAt:         o.CreatedAt,
	}
	newOrder.Delivery = DeliveryInResponse{
		OrderUid: o.OrderUID,
		Name:     o.Delivery.Name,
		Phone:    o.Delivery.Phone,
		Zip:      o.Delivery.Zip,
		City:     o.Delivery.City,
		Address:  o.Delivery.Address,
		Region:   o.Delivery.Region,
		Email:    o.Delivery.Email,
	}
	newOrder.Payment = PaymentInResponse{
		OrderUid:     o.OrderUID,
		Transaction:  o.Payment.Transaction,
		RequestId:    o.Payment.RequestId,
		Currency:     o.Payment.Currency,
		Provider:     o.Payment.Provider,
		Amount:       o.Payment.Amount,
		PaymentDt:    o.Payment.PaymentDt,
		Bank:         o.Payment.Bank,
		DeliveryCost: o.Payment.DeliveryCost,
		GoodsTotal:   o.Payment.GoodsTotal,
		CustomFee:    o.Payment.CustomFee,
	}
	for _, item := range o.Items {
		newOrder.Items = append(newOrder.Items, ItemInResponse{
			OrderUid:    item.OrderUid,
			ChrtId:      item.ChrtId,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			Rid:         item.Rid,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmId:        item.NmId,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}
	return newOrder
}
