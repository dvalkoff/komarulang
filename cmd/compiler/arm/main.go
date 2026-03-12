package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dvalkoff/komarulang/codegen"
	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
)

const (
	outputFile = "output"
	inputFile  = "input"
)

func getCompilerParameters() map[string]string {
	parameters := map[string]string{}
	args := os.Args
	for i := 0; i < len(args)-1; {
		arg := args[i]
		switch arg {
		case "-o":
			i++
			parameters[outputFile] = args[i]
		}
		i++
	}
	parameters[inputFile] = args[len(args)-1]
	if _, ok := parameters[outputFile]; !ok {
		parameters[outputFile] = "out"
	}
	return parameters
}

func assemble(asmFile, objFile string) error {
	cmd := exec.Command("as", "-o", objFile, asmFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func link(objFile, outFile string) error {
	xcrun := exec.Command("xcrun", "-sdk", "macosx", "--show-sdk-path")
	sdkPathBytes, err := xcrun.Output()
	if err != nil {
		return fmt.Errorf("xcrun failed: %w", err)
	}
	sdkPath := strings.TrimSpace(string(sdkPathBytes))

	cmd := exec.Command("ld",
		"-o", outFile,
		objFile,
		"-lSystem",
		"-syslibroot", sdkPath,
		"-e", "_main",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	params := getCompilerParameters()
	tokenizer := tokenizer.Tokenizer{File: params[inputFile]}
	sourceFile, err := os.Open(params[inputFile])
	if err != nil {
		panic("Failed to open source file")
	}
	defer sourceFile.Close()
	tokens, err := tokenizer.Scan(sourceFile)
	if err != nil {
		panic(err)
	}
	p := parser.NewParser(tokens)
	ast, err := p.Expression()
	if err != nil {
		panic(err)
	}
	codegen := codegen.NewCodeGenerator()
	if _, err := codegen.CompileExpr(ast, 0); err != nil {
		panic(err)
	}
	as := codegen.Prog.String()

	assemblyFileName := fmt.Sprintf("%v.s", params[outputFile])
	objectFileName := fmt.Sprintf("%v.o", params[outputFile])
	binaryFileName := params[outputFile]

	assemblyFile, err := os.Create(assemblyFileName)
	if err != nil {
		panic(err)
	}
	defer assemblyFile.Close()
	writer := bufio.NewWriter(assemblyFile)
	_, err = writer.WriteString(as)
	if err != nil {
		panic(err)
	}
	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	if err := assemble(assemblyFileName, objectFileName); err != nil {
		panic(fmt.Sprintf("assembly failed: %v\n", err))
	}

	if err := link(objectFileName, binaryFileName); err != nil {
		panic(fmt.Sprintf("linking failed: %v\n", err))
	}

	fmt.Printf("Build successful: ./%v\n", binaryFileName)
}
