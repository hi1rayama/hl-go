package main

import (
	"fmt"
	"io"
	"os"
)

// type ProgramCode struct {
// 	text []byte
// }

func loadText(name string) ([]byte, error) {
	// func (pc *ProgramCode) loadText(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b := make([]byte, 1024)
	for {
		c, err := file.Read(b)
		if c == 0 {
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	// pc.text = b

	// fmt.Println(pc.text)
	return b, nil
}

func main() {
	// program := ProgramCode{}
	variable := make([]int, 256)
	var i int
	txt, err := loadText("test.txt")
	if err != nil {
		fmt.Println(err)
	}

	for i = 0; i < 10; i++ {
		variable['0'+i] = i // '0'はasciiで48, つまりindexのascci文字コードと数値が対応している
		fmt.Println('0'+i, variable['0'+i])
	}

	for pc := 0; txt[pc] != 0; pc++ {

		// 改行や文終端文字を飛ばす
		if txt[pc] == '\n' || txt[pc] == '\r' || txt[pc] == '\t' || txt[pc] == ';' {
			continue
		}

		if txt[pc+1] == '=' && txt[pc+3] == ';' { // 単純に代入  変数(0) =(1) 値(2) ;(3)
			variable[txt[pc]] = variable[txt[pc+2]]
		} else if txt[pc+1] == '=' && txt[pc+3] == '+' && txt[pc+5] == ';' { // 加算 変数(0) =(1) 値(2) +(3) 値(4) ;(5)
			variable[txt[pc]] = variable[txt[pc+2]] + variable[txt[pc+4]]
		} else if txt[pc+1] == '=' && txt[pc+3] == '-' && txt[pc+5] == ';' { // 加算 変数(0) =(1) 値(2) -(3) 値(4) ;(5)
			variable[txt[pc]] = variable[txt[pc+2]] - variable[txt[pc+4]]
		} else if txt[pc] == 'p' && txt[pc+1] == 'r' && txt[pc+5] == ' ' && txt[pc+7] == ';' { // 最初の2文字しか調べてない
			fmt.Printf("%d\n", variable[txt[pc+6]])
		} else {
			fmt.Printf("syntax error : %.10s\n", &txt[pc])
			os.Exit(1)
		}

		// 終端文字までカウンタをインクリメントする
		for txt[pc] != ';' {
			pc++
		}

	}

}
