package cashaccount

type UserAmount struct {
	ID     uint32 `json:"id"`
	Amount uint32 `json:"amount"`
}

type UserID struct {
	ID uint32 `json:"id"`
}

type MoneyTransferDetails struct {
	FromId uint32 `json:"from_id"`
	ToId   uint32 `json:"to_id"`
	Amount uint32 `json:"amount"`
}

type ReserveDetails struct {
	ID        uint32 `json:"id"`
	ServiceId uint32 `json:"service_id"`
	OrderId   uint32 `json:"order_id"`
	Amount    uint32 `json:"amount"`
}
