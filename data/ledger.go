package data

type LedgerHeader struct {
	LedgerSequence  uint32     `json:"ledger_index,string"`
	TotalXRP        uint64     `json:"total_coins,string"`
	PreviousLedger  Hash256    `json:"parent_hash"`
	TransactionHash Hash256    `json:"transaction_hash"`
	StateHash       Hash256    `json:"account_hash"`
	ParentCloseTime RippleTime `json:"-"`
	CloseTime       RippleTime `json:"close_time"`
	CloseResolution uint8      `json:"close_time_resolution"`
	CloseFlags      uint8      `json:"-"`
}

type Ledger struct {
	hashable
	LedgerHeader
	Closed       bool             `json:"closed"`
	Accepted     bool             `json:"accepted"`
	Transactions TransactionSlice `json:"transactions"`
	AccountState LedgerEntrySlice `json:"accountState"`
}

func NewEmptyLedger(sequence uint32) *Ledger {
	return &Ledger{
		LedgerHeader: LedgerHeader{
			LedgerSequence: sequence,
		},
	}
}

func (l *Ledger) GetType() string {
	return "LedgerMaster"
}
