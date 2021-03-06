package data

// Evil things happen here. Rippled needs a V2 API...

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

type ledgerJSON Ledger

// adds all the legacy fields
type ledgerExtraJSON struct {
	ledgerJSON
	HumanCloseTime *RippleHumanTime `json:"close_time_human"`
	Hash           Hash256          `json:"hash"`
	LedgerHash     Hash256          `json:"ledger_hash"`
	TotalCoins     uint64           `json:"totalCoins,string"`
	SequenceNumber uint32           `json:"seqNum,string"`
}

func (l Ledger) MarshalJSON() ([]byte, error) {
	return json.Marshal(ledgerExtraJSON{
		ledgerJSON:     ledgerJSON(l),
		HumanCloseTime: l.CloseTime.Human(),
		Hash:           l.Hash(),
		LedgerHash:     l.Hash(),
		TotalCoins:     l.TotalXRP,
		SequenceNumber: l.LedgerSequence,
	})
}

func (l *Ledger) UnmarshalJSON(b []byte) error {
	var ledger ledgerExtraJSON
	if err := json.Unmarshal(b, &ledger); err != nil {
		return err
	}
	*l = Ledger(ledger.ledgerJSON)
	l.SetHash(ledger.Hash[:])
	return nil
}

// Wrapper types to enable second level of marshalling
// when found in ledger API call
type txmLedger struct {
	MetaData MetaData `json:"metaData"`
}

// Wrapper types to enable second level of marshalling
// when found in tx API call
type txmNormal TransactionWithMetaData

var (
	txmTransactionTypeRegex = regexp.MustCompile(`"TransactionType":.*?"(.*?)"`)
	txmHashRegex            = regexp.MustCompile(`"hash":.*?"(.*?)"`)
	txmMetaTypeRegex        = regexp.MustCompile(`"(meta|metaData)"`)
)

func (txm *TransactionWithMetaData) UnmarshalJSON(b []byte) error {
	txTypeMatch := txmTransactionTypeRegex.FindStringSubmatch(string(b))
	hashMatch := txmHashRegex.FindStringSubmatch(string(b))
	metaTypeMatch := txmMetaTypeRegex.FindStringSubmatch(string(b))
	var txType, hash, metaType string
	if txTypeMatch == nil {
		return fmt.Errorf("Not a valid transaction with metadata: Missing TransactionType")
	}
	txType = txTypeMatch[1]
	if hashMatch == nil {
		return fmt.Errorf("Not a valid transaction with metadata: Missing Hash")
	}
	hash = hashMatch[1]
	if metaTypeMatch != nil {
		metaType = metaTypeMatch[1]
	}
	txm.Transaction = GetTxFactoryByType(txType)()
	h, err := hex.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("Bad hash: %s", hash)
	}
	txm.SetHash(h)
	if err := json.Unmarshal(b, txm.Transaction); err != nil {
		return err
	}
	switch metaType {
	case "meta":
		return json.Unmarshal(b, (*txmNormal)(txm))
	case "metaData":
		var meta txmLedger
		if err := json.Unmarshal(b, &meta); err != nil {
			return err
		}
		txm.MetaData = meta.MetaData
		return nil
	default:
		return nil
	}
}

func (txm TransactionWithMetaData) marshalJSON() ([]byte, []byte, error) {
	tx, err := json.Marshal(txm.Transaction)
	if err != nil {
		return nil, nil, err
	}
	meta, err := json.Marshal(txm.MetaData)
	if err != nil {
		return nil, nil, err
	}
	return tx, meta, nil
}

type extractTxm struct {
	Tx   json.RawMessage `json:"transaction"`
	Meta json.RawMessage `json:"meta"`
}

const extractTxmFormat = `%s,"meta":%s}`

func UnmarshalTransactionWithMetadata(b []byte, txm *TransactionWithMetaData) error {
	var extract extractTxm
	if err := json.Unmarshal(b, &extract); err != nil {
		return err
	}
	raw := fmt.Sprintf(extractTxmFormat, extract.Tx[:len(extract.Tx)-1], extract.Meta)
	return json.Unmarshal([]byte(raw), txm)
}

const txmFormat = `%s,"hash":"%s","inLedger":%d,"ledger_index":%d,"meta":%s}`

func (txm TransactionWithMetaData) MarshalJSON() ([]byte, error) {
	tx, meta, err := txm.marshalJSON()
	if err != nil {
		return nil, err
	}
	out := fmt.Sprintf(txmFormat, string(tx[:len(tx)-1]), txm.Hash().String(), txm.LedgerSequence, txm.LedgerSequence, string(meta))
	return []byte(out), nil
}

const txmSliceFormat = `%s,"hash":"%s","metaData":%s}`

func (s TransactionSlice) MarshalJSON() ([]byte, error) {
	raw := make([]json.RawMessage, len(s))
	var err error
	var tx, meta []byte
	for i, txm := range s {
		if tx, meta, err = txm.marshalJSON(); err != nil {
			return nil, err
		}
		extra := fmt.Sprintf(txmSliceFormat, string(tx[:len(tx)-1]), txm.Hash().String(), meta)
		raw[i] = json.RawMessage(extra)
	}
	return json.Marshal(raw)
}

var (
	leTypeRegex  = regexp.MustCompile(`"LedgerEntryType":.*"(.*)"`)
	leIndexRegex = regexp.MustCompile(`"index":.*"(.*)"`)
)

func (l *LedgerEntrySlice) UnmarshalJSON(b []byte) error {
	var s []json.RawMessage
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for _, raw := range s {
		leTypeMatch := leTypeRegex.FindStringSubmatch(string(raw))
		indexMatch := leIndexRegex.FindStringSubmatch(string(raw))
		if leTypeMatch == nil {
			return fmt.Errorf("Bad LedgerEntryType")
		}
		if indexMatch == nil {
			return fmt.Errorf("Missing LedgerEntry index")
		}
		le := GetLedgerEntryFactoryByType(leTypeMatch[1])()
		index, err := hex.DecodeString(indexMatch[1])
		if err != nil {
			return fmt.Errorf("Bad index: %s", index)
		}
		le.SetHash(index)
		if err := json.Unmarshal(raw, &le); err != nil {
			return err
		}
		*l = append(*l, le)
	}
	return nil
}

const leSliceFormat = `%s,"index":"%s"}`

func (s LedgerEntrySlice) MarshalJSON() ([]byte, error) {
	raw := make([]json.RawMessage, len(s))
	var err error
	for i, le := range s {
		if raw[i], err = json.Marshal(le); err != nil {
			return nil, err
		}
		extra := fmt.Sprintf(leSliceFormat, string(raw[i][:len(raw[i])-1]), le.Hash().String())
		raw[i] = json.RawMessage(extra)
	}
	return json.Marshal(raw)
}

func (i NodeIndex) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%016X", i)), nil
}

func (i *NodeIndex) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 16, 64)
	*i = NodeIndex(n)
	return err
}

func (r TransactionResult) MarshalText() ([]byte, error) {
	return []byte(resultNames[r]), nil
}

func (r *TransactionResult) UnmarshalText(b []byte) error {
	if result, ok := reverseResults[string(b)]; ok {
		*r = result
		return nil
	}
	return fmt.Errorf("Unknown TransactionResult: %s", string(b))
}

func (l LedgerEntryType) MarshalText() ([]byte, error) {
	return []byte(ledgerEntryNames[l]), nil
}

func (l *LedgerEntryType) UnmarshalText(b []byte) error {
	if leType, ok := ledgerEntryTypes[string(b)]; ok {
		*l = leType
		return nil
	}
	return fmt.Errorf("Unknown LedgerEntryType: %s", string(b))
}

func (t TransactionType) MarshalText() ([]byte, error) {
	return []byte(txNames[t]), nil
}

func (t *TransactionType) UnmarshalText(b []byte) error {
	if txType, ok := txTypes[string(b)]; ok {
		*t = txType
		return nil
	}
	return fmt.Errorf("Unknown TransactionType: %s", string(b))
}

func (t RippleTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(t.Uint32()), 10)), nil
}

func (t *RippleTime) UnmarshalJSON(b []byte) error {
	n, err := strconv.ParseUint(string(b), 10, 32)
	if err != nil {
		return err
	}
	t.SetUint32(uint32(n))
	return nil
}

func (t RippleHumanTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *RippleHumanTime) UnmarshalJSON(b []byte) error {
	t.RippleTime = &RippleTime{}
	return t.SetString(string(b[1 : len(b)-1]))
}

func (v *Value) MarshalText() ([]byte, error) {
	if v.Native {
		return []byte(strconv.FormatUint(v.Num, 10)), nil
	}
	return []byte(v.String()), nil
}

func (v *Value) UnmarshalText(b []byte) error {
	value, err := NewValue(string(b), true)
	if err != nil {
		return err
	}
	*v = *value
	return nil
}

type nonNativeValue Value

func (v *nonNativeValue) UnmarshalText(b []byte) error {
	value, err := NewValue(string(b), false)
	if err != nil {
		return err
	}
	*v = nonNativeValue(*value)
	return nil
	// return (*Value)(v).Parse(string(b))
}

func (v *nonNativeValue) MarshalText() ([]byte, error) {
	return (*Value)(v).MarshalText()
}

type amountJSON struct {
	Value    *nonNativeValue `json:"value"`
	Currency Currency        `json:"currency"`
	Issuer   Account         `json:"issuer"`
}

func (a Amount) MarshalJSON() ([]byte, error) {
	if a.Native {
		return []byte(`"` + strconv.FormatUint(a.Num, 10) + `"`), nil
	}
	return json.Marshal(amountJSON{(*nonNativeValue)(a.Value), a.Currency, a.Issuer})
}

func (a *Amount) UnmarshalJSON(b []byte) (err error) {
	if b[0] != '{' {
		a.Value = new(Value)
		return json.Unmarshal(b, a.Value)
	}
	var dummy amountJSON
	if err := json.Unmarshal(b, &dummy); err != nil {
		return err
	}
	a.Value, a.Currency, a.Issuer = (*Value)(dummy.Value), dummy.Currency, dummy.Issuer
	return nil
}

func (c Currency) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *Currency) UnmarshalText(text []byte) error {
	var err error
	*c, err = NewCurrency(string(text))
	return err
}

func (h Hash128) MarshalText() ([]byte, error) {
	return b2h(h[:]), nil
}

func (h *Hash128) UnmarshalText(b []byte) error {
	_, err := hex.Decode(h[:], b)
	return err
}

func (h Hash160) MarshalText() ([]byte, error) {
	return b2h(h[:]), nil
}

func (h *Hash160) UnmarshalText(b []byte) error {
	_, err := hex.Decode(h[:], b)
	return err
}

func (h Hash256) MarshalText() ([]byte, error) {
	return b2h(h[:]), nil
}

func (h *Hash256) UnmarshalText(b []byte) error {
	_, err := hex.Decode(h[:], b)
	return err
}

func (a Account) MarshalText() ([]byte, error) {
	address, err := a.Hash()
	if err != nil {
		return nil, err
	}
	return address.MarshalText()
}

// Expects base58-encoded account id
func (a *Account) UnmarshalText(b []byte) error {
	account, err := NewAccountFromAddress(string(b))
	if err != nil {
		return err
	}
	copy(a[:], account[:])
	return nil
}

func (r RegularKey) MarshalText() ([]byte, error) {
	address, err := r.Hash()
	if err != nil {
		return nil, err
	}
	return address.MarshalText()
}

// Expects base58-encoded account id
func (r *RegularKey) UnmarshalText(b []byte) error {
	account, err := NewRegularKeyFromAddress(string(b))
	if err != nil {
		return err
	}
	copy(r[:], account[:])
	return nil
}

func (v VariableLength) MarshalText() ([]byte, error) {
	return b2h(v), nil
}

// Expects variable length hex
func (v *VariableLength) UnmarshalText(b []byte) error {
	var err error
	*v, err = hex.DecodeString(string(b))
	return err
}

func (p PublicKey) MarshalText() ([]byte, error) {
	return b2h(p[:]), nil
}

// Expects public key hex
func (p *PublicKey) UnmarshalText(b []byte) error {
	_, err := hex.Decode(p[:], b)
	return err
}

type affectedNodeJSON AffectedNode

type affectedFields struct {
	*affectedNodeJSON
	FinalFields    *json.RawMessage
	PreviousFields *json.RawMessage
	NewFields      *json.RawMessage
}

func (n *AffectedNode) UnmarshalJSON(b []byte) error {
	affected := affectedFields{
		affectedNodeJSON: (*affectedNodeJSON)(n),
	}
	if err := json.Unmarshal(b, &affected); err != nil {
		return err
	}
	if affected.FinalFields != nil {
		n.FinalFields = FieldsFactory[n.LedgerEntryType]()
		if err := json.Unmarshal(*affected.FinalFields, n.FinalFields); err != nil {
			return err
		}
	}
	if affected.PreviousFields != nil {
		n.PreviousFields = FieldsFactory[n.LedgerEntryType]()
		if err := json.Unmarshal(*affected.PreviousFields, n.PreviousFields); err != nil {
			return err
		}
	}
	if affected.NewFields != nil {
		n.NewFields = FieldsFactory[n.LedgerEntryType]()
		if err := json.Unmarshal(*affected.NewFields, n.NewFields); err != nil {
			return err
		}
	}
	return nil
}
