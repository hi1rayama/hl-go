package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const MAX_TC = 255 // トークンコードの最大値

// C言語のenumを表現
const (
	TcSemi   int = iota // 0
	TcDot               // 1
	TcWiCard            // 2
	Tc0                 // 3
	Tc1
	Tc2
	Tc3
	Tc4
	Tc5
	Tc6
	Tc7
	Tc8
	TcEEq
	TcNEq
	TcLt
	TcGe
	TcLe
	TcGt
)

var ts = make([]string, MAX_TC+1)        // トークンの内容を記憶
var tl = make([]int, MAX_TC+1)           // トークンの長さ
var tcBuff = make([]byte, (MAX_TC+1)*10) // トークン１つ当たり平均10バイトを想定.
var tcs = 0
var tcb = 0

var tc = make([]int, 10000)

var variable = make([]int, MAX_TC+1)

var tcInit = []byte("; . !!* 0 1 2 3 4 5 6 7 8 == != < >= <= >")

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
		// fmt.Printf("j = %d [tc[j] = %d : ts[tc[j]] = %s]\n", j, tc[j], ts[tc[j]])
		i += len
		j++

	}

}

var phrCmp_tc = make([]int, 32*100)
var ppc1 int
var wpc = make([]int, 9)

// phrCmp phrで指定されたトークン列と一致するかどうか調べる
func phrCmp(pid int, phr string, pc int) bool {

	// 文字列をバイト列に変換
	phrb := []byte(phr)
	phrb = append(phrb, 0)

	i0 := pid * 32
	var i, i1, j int

	if phrCmp_tc[i0+31] == 0 {
		i1 = lexer(phrb, phrCmp_tc[i0:])
		phrCmp_tc[i0+31] = i1
	}
	i1 = phrCmp_tc[i0+31]
	for i = 0; i < i1; i++ {
		if phrCmp_tc[i0+i] == TcWiCard {
			i++
			j = phrCmp_tc[i0+i] - Tc0
			wpc[j] = pc
			pc++
			continue
		}
		if phrCmp_tc[i0+i] != tc[pc] {
			return false
		}
		pc++
	}
	ppc1 = pc
	return true
}

// run 言語処理本体
func run(s []byte) {
	start := time.Now()
	// 字句解析を行い, トークンの数を取得
	pc1 := lexer(s, tc)
	tc[pc1] = TcSemi // 末尾に「;」を付け忘れることが多いので、付けてあげる.
	pc1++

	// エラー表示用のために末尾にピリオドを登録しておく.
	tc[pc1] = TcDot
	tc[pc1+1] = TcDot
	tc[pc1+2] = TcDot
	tc[pc1+3] = TcDot

	// goto文を実現するためにラベル定義命令の次のPC値を記憶させる
	for pc := 0; pc < pc1; pc++ {
		if phrCmp(0, "!!*0:", pc) {
			variable[tc[pc]] = ppc1 // ラベル定義命令の次のpc値を変数に記憶させておく
		}
	}

	// 構文解析を行う
	for pc := 0; pc < pc1; {

		if phrCmp(1, "!!*0 = !!*1;", pc) { // 単純に代入  変数(0) =(1) 値(2) ;(3)
			variable[tc[wpc[0]]] = variable[tc[wpc[1]]]
		} else if phrCmp(2, "!!*0 = !!*1 + !!*2;", pc) { // 加算 変数(0) =(1) 値(2) +(3) 値(4) ;(5)
			variable[tc[wpc[0]]] = variable[tc[wpc[1]]] + variable[tc[wpc[2]]]
		} else if phrCmp(3, "!!*0 = !!*1 - !!*2;", pc) { // 加算 変数(0) =(1) 値(2) -(3) 値(4) ;(5)
			variable[tc[wpc[0]]] = variable[tc[wpc[2]]] - variable[tc[wpc[2]]]
		} else if phrCmp(4, "print !!*0;", pc) {
			fmt.Printf("%d\n", variable[tc[wpc[0]]])
		} else if phrCmp(0, "!!*0:", pc) { // ラベル定義命令
			// continue
		} else if phrCmp(5, "goto !!*0;", pc) { //goto
			pc = variable[tc[wpc[0]]]
			continue
		} else if phrCmp(6, "if (!!*0 !!*1 !!*2) goto !!*3;", pc) && TcEEq <= tc[wpc[1]] && tc[wpc[1]] <= TcGt {
			gpc := variable[tc[wpc[3]]]
			v0 := variable[tc[wpc[0]]]
			cc := tc[wpc[1]]
			v1 := variable[tc[wpc[2]]]
			// 条件が成立したらgoto処理.
			if cc == TcEEq && v0 != v1 {
				pc = gpc
				continue
			}
			if cc == TcNEq && v0 == v1 {
				pc = gpc
				continue
			}
			if cc == TcLt && v0 < v1 {
				pc = gpc
				continue
			}
		} else if phrCmp(7, "time;", pc) {
			end := time.Now()
			fmt.Printf("%f[sec]\n", (end.Sub(start)).Seconds())
		} else if phrCmp(8, ";", pc) {
			// 何もしない
		} else {
			fmt.Printf("syntax error : %s %s %s %s\n", ts[tc[pc]], ts[tc[pc+1]], ts[tc[pc+2]], ts[tc[pc+3]])
			os.Exit(1)
		}

		pc = ppc1
	}

}

func main() {
	tcInit = append(tcInit, 0)
	lexer(tcInit, tc)
	// コマンドライン引数のチェック
	flag.Parse()
	args := flag.Args()
	if len(args) >= 1 {
		program, err := loadText(args[0])
		if err != nil {
			fmt.Printf("fopen error : %s\n", args[0])
			os.Exit(1)
		}
		run(program)
		os.Exit(0)
	}

	// REPL
	stdin := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\n>")
		stdin.Scan()
		txt := []byte(stdin.Text())
		fmt.Println(txt)
		i := len(txt)
		if txt[i-1] == '\n' { // 末尾に改行コードが付いていればそれを消す.
			txt[i-1] = 0
		}

		// ファイルの実行
		if string(txt[:4]) == "run " {
			program, err := loadText(string(txt[4:]))
			if err != nil {
				fmt.Printf("fopen error : %s\n", string(txt[4:]))
			} else {
				run(program)
			}
		} else if string(txt[:4]) == "exit" {
			os.Exit(0)
		} else {
			txt = append(txt, 0)
			run(txt)
		}

	}
}
