package cashaccount

import "time"

type UserAmount struct {
	ID     uint32  `json:"id"`
	Amount float32 `json:"amount"`
}

type MoneyTransferDetails struct {
	FromId uint32  `json:"from_id"`
	ToId   uint32  `json:"to_id"`
	Amount float32 `json:"amount"`
}

type ReserveDetails struct {
	ID        uint32  `json:"id"`
	ServiceId uint32  `json:"service_id"`
	OrderId   uint32  `json:"order_id"`
	Amount    float32 `json:"amount"`
}

type TextResponse struct {
	Message string `json:"message"`
}

type ReportLinkResponse struct {
	Link string `json:"link"`
}

type BookkeepingReportRow struct {
	ServiceId uint32
	Amount    float32
}

type UserReportRow struct {
	Amount      float32   `json:"amount"`
	Description string    `json:"description"`
	DateTime    time.Time `json:"dateTime"`
}
