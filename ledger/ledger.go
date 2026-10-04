package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
)

// В этом примере каждая транзакция — расход с положительной суммой.
type Transaction struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

// Budget задаёт постоянный лимит категории, без разделения по периодам.
type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
}

var ErrBudgetExceeded = errors.New("budget exceeded")

func NewLedger() *Ledger {
	return &Ledger{budgets: make(map[string]Budget)}
}

// SetBudget добавляет бюджет или заменяет существующий бюджет категории.
// Уменьшение лимита не удаляет уже записанные транзакции.
func (l *Ledger) SetBudget(b Budget) {
	if l.budgets == nil {
		l.budgets = make(map[string]Budget)
	}
	l.budgets[b.Category] = b
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	if tx.Category == "" || tx.Amount <= 0 || math.IsNaN(tx.Amount) || math.IsInf(tx.Amount, 0) {
		return errors.New("transaction must have a category and a finite positive amount")
	}
	if b, ok := l.budgets[tx.Category]; ok {
		if err := validateBudget(b); err != nil {
			return err
		}
		total := tx.Amount
		for _, existing := range l.transactions {
			if existing.Category == tx.Category {
				total += existing.Amount
			}
		}
		if total > b.Limit {
			return fmt.Errorf("%w: category %q, total %.2f, limit %.2f", ErrBudgetExceeded, tx.Category, total, b.Limit)
		}
	}
	l.transactions = append(l.transactions, tx)
	return nil
}

// Transactions возвращает копию, чтобы вызывающий код не менял учёт расходов.
func (l *Ledger) Transactions() []Transaction {
	return append([]Transaction(nil), l.transactions...)
}

func validateBudget(b Budget) error {
	if b.Category == "" || b.Limit < 0 || math.IsNaN(b.Limit) || math.IsInf(b.Limit, 0) {
		return fmt.Errorf("invalid budget for category %q: expected a category and a finite non-negative limit", b.Category)
	}
	return nil
}

// LoadBudgets принимает один JSON-массив. При ошибке хранилище не меняется.
func (l *Ledger) LoadBudgets(r io.Reader) error {
	decoder := json.NewDecoder(r)
	var budgets []Budget
	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("read or parse budgets JSON: %w", err)
	}
	if budgets == nil {
		return errors.New("budgets JSON must be an array, not null")
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return fmt.Errorf("read trailing budgets JSON: %w", err)
		}
		return errors.New("budgets JSON must contain exactly one array")
	}
	for i, b := range budgets {
		if err := validateBudget(b); err != nil {
			return fmt.Errorf("budget at index %d: %w", i, err)
		}
	}
	for _, b := range budgets {
		l.SetBudget(b)
	}
	return nil
}
