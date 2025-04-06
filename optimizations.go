package wav2ulaw

import (
	"math"
	"sync"
)

const (
	// Розмір таблиці для кожного вікна
	tableSize = 1024
)

// SincTable зберігає попередньо обчислені значення sinc функції
type SincTable struct {
	windowSize int
	values     []float64
}

var (
	// Кеш таблиць для різних розмірів вікна
	sincTableCache = make(map[int]*SincTable)
	cacheMutex     sync.RWMutex
)

// getSincTable повертає або створює таблицю sinc значень для заданого розміру вікна
func getSincTable(windowSize int) *SincTable {
	cacheMutex.RLock()
	table, exists := sincTableCache[windowSize]
	cacheMutex.RUnlock()

	if exists {
		return table
	}

	// Якщо таблиця не існує, створюємо нову
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Перевіряємо ще раз після отримання блокування
	if table, exists = sincTableCache[windowSize]; exists {
		return table
	}

	table = &SincTable{
		windowSize: windowSize,
		values:     make([]float64, tableSize),
	}

	// Обчислюємо значення sinc для різних x
	for i := 0; i < tableSize; i++ {
		x := float64(i) / float64(tableSize-1) * float64(windowSize) * math.Pi
		if x == 0 {
			table.values[i] = 1.0
		} else {
			table.values[i] = math.Sin(x) / x
		}
	}

	sincTableCache[windowSize] = table
	return table
}

// getSincValue повертає інтерпольоване значення sinc з таблиці
func (t *SincTable) getSincValue(x float64) float64 {
	// Нормалізуємо x до діапазону таблиці
	x = math.Abs(x)
	if x >= float64(t.windowSize)*math.Pi {
		return 0
	}

	// Знаходимо індекс в таблиці
	idx := x * float64(tableSize-1) / (float64(t.windowSize) * math.Pi)
	i := int(idx)
	
	// Лінійна інтерполяція між сусідніми значеннями
	if i >= tableSize-1 {
		return t.values[tableSize-1]
	}
	
	frac := idx - float64(i)
	return t.values[i]*(1-frac) + t.values[i+1]*frac
}

// Оновлена версія resamplePCM16 з використанням попередньо обчисленої таблиці
func resamplePCM16WithTable(input []int16, inputRate, outputRate float64, windowSize int) []int16 {
	ratio := outputRate / inputRate
	outputLen := int(float64(len(input)) * ratio)
	output := make([]int16, outputLen)

	// Отримуємо таблицю sinc значень
	sincTable := getSincTable(windowSize)

	// Pre-calculate window coefficients
	window := make([]float64, windowSize*2+1)
	for i := range window {
		// Blackman window
		x := float64(i) / float64(len(window)-1)
		window[i] = 0.42 - 0.5*math.Cos(2*math.Pi*x) + 0.08*math.Cos(4*math.Pi*x)
	}

	for i := range output {
		pos := float64(i) / ratio
		idx := int(pos)
		
		sum := 0.0
		weightSum := 0.0

		for j := -windowSize; j <= windowSize; j++ {
			inputIdx := idx + j
			if inputIdx < 0 || inputIdx >= len(input) {
				continue
			}

			// Використовуємо попередньо обчислене значення sinc
			x := math.Pi * (pos - float64(inputIdx))
			sinc := sincTable.getSincValue(x)

			// Apply window function
			weight := window[j+windowSize] * sinc
			sum += float64(input[inputIdx]) * weight
			weightSum += weight
		}

		if weightSum > 0 {
			sum /= weightSum
		}
		output[i] = int16(math.Round(sum))
	}

	return output
} 