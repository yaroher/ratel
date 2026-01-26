package repository

// Converter - функции конвертации между Scanner и Proto типами
type Converter[S, P any] struct {
	ToScanner func(P) S
	ToProto   func(S) P
}
