package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	ledger := NewLedger()
	ledger.SetBudget(Budget{Category: "еда", Limit: 5000})
	ledger.SetBudget(Budget{Category: "транспорт", Limit: 2000})

	file, err := os.Open("budgets.json")
	if err != nil {
		log.Fatalf("Не удалось открыть budgets.json: %v", err)
	}
	defer file.Close()
	if err := ledger.LoadBudgets(bufio.NewReader(file)); err != nil {
		log.Fatalf("Не удалось загрузить бюджеты: %v", err)
	}

	for _, tx := range []Transaction{
		{Category: "еда", Amount: 3000}, // успех: 3000 из 5000
		{Category: "еда", Amount: 2000}, // успех: ровно 5000
		{Category: "еда", Amount: 100},  // отказ: 5100 > 5000
		{Category: "транспорт", Amount: 500},
		{Category: "книги", Amount: 700}, // категория без бюджета
	} {
		if err := ledger.AddTransaction(tx); err != nil {
			fmt.Printf("Отказ: %s, %.2f руб.: %v\n", tx.Category, tx.Amount, err)
			continue
		}
		fmt.Printf("Добавлено: %s, %.2f руб.\n", tx.Category, tx.Amount)
	}

	fmt.Printf("Сохранено транзакций: %d\n", len(ledger.Transactions()))
	for _, tx := range ledger.Transactions() {
		fmt.Printf("  %s: %.2f руб.\n", tx.Category, tx.Amount)
	}
}
