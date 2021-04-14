package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const MAX_TC = 1000 // トークンコードの最大値

var ts = make([]string, MAX_TC+1)        // トークンの内容を記憶
var tl = make([]int, MAX_TC+1)           // トークンの長さ
var tcBuff = make([]byte, (MAX_TC+1)*10) // トークン１つ当たり平均10バイトを想定.
var tcs = 0
var tcb = 0

var variable = make([]int, MAX_TC+1)

// loadText ファイルを読み込む
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

// getTc トークンの登録と登録されているトークンの位置(インデックス)を返す
func getTc(s string, len int) int {
	i := 0
	// トークンが登録されているかを確認
	for i = 0; i < tcs; i++ {
		if len == tl[i] && ts[i] == s[:len] {
			break
		}
	}
	// トークンが登録されていなかった場合はトークンを登録する
	if i == tcs {
		// 登録されているトークンが最大だった場合はエラー
		if tcs >= MAX_TC {
			fmt.Printf("too many tokens\n")
			os.Exit(1)
		}

		tcBuff = []byte(s[:len])
		ts[i] = string(tcBuff)
		tl[i] = len
		tcb += len + 1
		tcs++

		// 定数の登録
		num, err := strconv.Atoi(ts[i])
		if err != nil {
			variable[i] = 0
		} else {
			variable[i] = num
		}
	}
	return i
}

// isAlpnabetOrNumber アルファベットか数字かを判断する
func isAlpnabetOrNumber(c byte) bool {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_' {
		return true
	}
	return false
}

// lexer 字句解析を行う関数
func lexer(s []byte, tc []int) int {
	i := 0
	j := 0

	for {
		// 改行や文終端文字を飛ばす
		if s[i] == '\n' || s[i] == '\r' || s[i] == '\t' || s[i] == ' ' {
			i++
			continue
		}
		// fmt.Println(string(s[i]))
		// プログラム終端まできたら, 解析終了
		if s[i] == 0 {
			return j
		}
		len := 0
		// 1文字目が記号かどうか
		if strings.Contains("(){}[];,", string(s[i])) {
			len = 1
		} else if isAlpnabetOrNumber(s[i]) { // アルファベットか数字だった場合, 読み進める
			for isAlpnabetOrNumber(s[i+len]) {
				len++
			}
		} else if strings.Contains("=+-*/!%&~|<>?:.#", string(s[i])) {
			for strings.Contains("=+-*/!%&~|<>?:.#", string(s[i+len])) && s[i+len] != 0 { // 数字か文字列の字句を読み込む
				len++
			}
		} else {
			fmt.Printf("syntax error : %s\n", &s[i])
			os.Exit(1)

		}

		tc[j] = getTc(string(s[i:]), len) // トークンの登録を行う
		fmt.Printf("j = %d [tc[j] = %d : ts[tc[j]] = %s]\n", j, tc[j], ts[tc[j]])
		i += len
		j++

	}

}

var tc = make([]int, 10000)

func main() {

	start := time.Now()

	// コードの読み込み
	txt, err := loadText("test.txt")
	if err != nil {
		fmt.Println(err)
	}
	// 字句解析を行い, トークンの数を取得
	pc1 := lexer(txt, tc)
	fmt.Printf("pc1 : %d\n", pc1)

	// エラー表示用のために末尾にピリオドを登録しておく.
	tc[pc1] = getTc(".", 1)
	tc[pc1+1] = getTc(".", 1)
	tc[pc1+2] = getTc(".", 1)
	tc[pc1+3] = getTc(".", 1)
	semi := getTc(";", 1)

	// goto文を実現するためにラベル定義命令の次のPC値を記憶させる
	for pc := 0; pc < pc1; pc++ {
		if tc[pc+1] == getTc(":", 1) {
			variable[tc[pc]] = pc + 2 // ラベル定義命令の次のpc値を変数に記憶させておく
		}
	}

	// 構文解析を行う
	for pc := 0; pc < pc1; {

		if tc[pc+1] == getTc("=", 1) && tc[pc+3] == semi { // 単純に代入  変数(0) =(1) 値(2) ;(3)
			variable[tc[pc]] = variable[tc[pc+2]]
		} else if tc[pc+1] == getTc("=", 1) && tc[pc+3] == getTc("+", 1) && tc[pc+5] == semi { // 加算 変数(0) =(1) 値(2) +(3) 値(4) ;(5)
			variable[tc[pc]] = variable[tc[pc+2]] + variable[tc[pc+4]]
		} else if tc[pc+1] == getTc("=", 1) && tc[pc+3] == getTc("-", 1) && tc[pc+5] == semi { // 加算 変数(0) =(1) 値(2) -(3) 値(4) ;(5)
			variable[tc[pc]] = variable[tc[pc+2]] - variable[tc[pc+4]]
		} else if tc[pc] == getTc("print", 5) && tc[pc+2] == semi {
			fmt.Printf("%d\n", variable[tc[pc+1]])
		} else if tc[pc+1] == getTc(":", 1) {
			pc += 2
			continue
		} else if tc[pc] == getTc("goto", 4) && tc[pc+2] == semi {
			pc = variable[tc[pc+1]]
			continue
		} else if tc[pc] == getTc("if", 2) && tc[pc+1] == getTc("(", 1) && tc[pc+5] == getTc(")", 1) && tc[pc+6] == getTc("goto", 4) && tc[pc+8] == semi {
			gpc := variable[tc[pc+7]]
			v0 := variable[tc[pc+2]]
			v1 := variable[tc[pc+4]]
			// 条件が成立したらgoto処理.
			if tc[pc+3] == getTc("!=", 2) && v0 != v1 {
				pc = gpc
				continue
			}
			if tc[pc+3] == getTc("==", 2) && v0 == v1 {
				pc = gpc
				continue
			}
			if tc[pc+3] == getTc("<", 1) && v0 < v1 {
				pc = gpc
				continue
			}
		} else if tc[pc] == getTc("time", 4) && tc[pc+1] == semi {
			end := time.Now()
			fmt.Printf("%f[sec]\n", (end.Sub(start)).Seconds())
		} else {
			fmt.Printf("syntax error : %s %s %s %s\n", ts[tc[pc]], ts[tc[pc+1]], ts[tc[pc+2]], ts[tc[pc+3]])
			os.Exit(1)
		}

		// 終端文字までカウンタをインクリメントする
		for tc[pc] != semi {
			pc++
		}
		pc++
	}

}
