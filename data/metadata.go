package data

type AffectedNode struct {
	FinalFields       interface{}     `json:",omitempty"`
	LedgerEntryType   LedgerEntryType `json:",omitempty"`
	LedgerIndex       *Hash256        `json:",omitempty"`
	PreviousFields    interface{}     `json:",omitempty"`
	NewFields         interface{}     `json:",omitempty"`
	PreviousTxnID     *Hash256        `json:",omitempty"`
	PreviousTxnLgrSeq *uint32         `json:",omitempty"`
}

type NodeEffect struct {
	ModifiedNode *AffectedNode `json:",omitempty"`
	CreatedNode  *AffectedNode `json:",omitempty"`
	DeletedNode  *AffectedNode `json:",omitempty"`
}

type NodeEffects []NodeEffect

type MetaData struct {
	AffectedNodes     NodeEffects
	TransactionIndex  uint32
	TransactionResult TransactionResult
	DeliveredAmount   *Amount `json:",omitempty"`
}

type TransactionSlice []*TransactionWithMetaData

type TransactionWithMetaData struct {
	Transaction
	MetaData       MetaData `json:"meta"`
	LedgerSequence uint32   `json:"ledger_index"`
}

func (m *MetaData) GetType() string { return "Metadata" }

func (m *NodeEffect) Action() string {
	switch {
	case m.ModifiedNode != nil:
		return "Modified"
	case m.DeletedNode != nil:
		return "Deleted"
	default:
		return "Created"
	}
}
