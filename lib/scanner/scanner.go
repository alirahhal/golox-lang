package scanner

type Scanner struct {
	Start   string
	Current string
	Line    int
}

func New() *Scanner {
	scanner := new(Scanner)
	return scanner
}

func (scanner *Scanner) InitScanner(source string) {
	scanner.Start = source
	scanner.Current = source
	scanner.Line = 1
}
