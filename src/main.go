package main

import (
	. "upt/core"
	"upt/pipelines"
	"upt/testing"

	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var lexemes = flag.Bool("lex", false, "processa um arquivo e retorna os elementos lexicos")
var ast = flag.Bool("ast", false, "processa um arquivo e retorma uma arvore sintatica abstrata")
var mod = flag.Bool("mod", false, "processa um arquivo e retorma um módulo tipado")
var C = flag.Bool("C", false, "processa um arquivo e emite C")

var test = flag.Bool("test", false, "roda testes para todos os arquivos em uma pasta")

var verbose = flag.Bool("v", false, "testes verbosos")

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		Fatal("número de argumentos invalido\n")
	}
	eval(args[0])
}

func eval(filename string) {
	checkValid()
	if *test {
		res := Test(filename)
		printResults(res)
		return
	}
	normalMode(filename)
}

func normalMode(filename string) {
	switch true {
	case *lexemes:
		lexemes, err := pipelines.Lexemes(filename)
		Check(err)
		output := []string{}
		for _, lexeme := range lexemes {
			output = append(output, lexeme.Text)
		}
		fmt.Println(strings.Join(output, ", "))
	case *ast:
		n, err := pipelines.Ast(filename)
		Check(err)
		fmt.Println(n)
	case *mod:
		m, err := pipelines.Mod(filename)
		Check(err)
		fmt.Println(m.String())
	case *C:
		str, err := pipelines.GenC(filename)
		Check(err)
		fmt.Println(str)
	default:
		_, err := pipelines.Compile(filename)
		Check(err)
	}
}

func checkValid() {
	var selected = []bool{*lexemes, *ast, *mod, *C}
	var count = 0
	for _, b := range selected {
		if b {
			count++
		}
	}
	if count > 1 {
		Fatal("escolha apenas uma das seguintes flags: lex, ast, mod ou C")
	}
}

func Test(folder string) []*testing.TestResult {
	entries, err := os.ReadDir(folder)
	if err != nil {
		Fatal(err.Error() + "\n")
	}
	results := []*testing.TestResult{}
	for _, v := range entries {
		fullpath := folder + "/" + v.Name()
		if v.IsDir() {
			if *verbose {
				fmt.Print("\u001b[35m entering: " + fullpath + "\u001b[0m\n")
			}
			res := Test(fullpath)
			results = append(results, res...)
			if *verbose {
				fmt.Print("\u001b[35m leaving: " + fullpath + "\u001b[0m\n")
			}
		} else {
			res := testing.Test(fullpath)
			results = append(results, &res)
			if *verbose {
				fmt.Print(fullpath + "\t")
				fmt.Print(res.String() + "\n")
			}
		}
	}
	return results
}

func printResults(results []*testing.TestResult) {
	failed := 0
	fmt.Print("\n")
	for _, res := range results {
		if !res.Ok && res.Message != "" {
			fmt.Print(res.File + "\t" + res.Message + "\n")
		}
		if !res.Ok {
			failed += 1
		}
	}
	fmt.Print("\n")
	fmt.Print("falharam: " + strconv.Itoa(failed) + "\n")
	fmt.Print("total: " + strconv.Itoa(len(results)) + "\n")
}

func Check(e *Error) {
	if e != nil {
		Fatal(e.String() + "\n")
	}
}

func Fatal(s string) {
	os.Stderr.Write([]byte(s))
	os.Exit(0)
}
