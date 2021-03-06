package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("Добро пожаловать в comic - программу компиляции comi-файлов")
	inFilename, outFilename, err := filenamesFromCommandLine()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var inFile, outFile *os.File

	if inFilename != "" {
		if inFile, err = os.Open(inFilename); err != nil {
			log.Fatal(err)
		}
		defer inFile.Close()
	} else {
		os.Exit(0)
	}

	if outFilename != "" {
		if outFile, err = os.Create(outFilename); err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()
	} else {
		os.Exit(0)
	}

	if err = compile(inFile, outFile); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nКомпиляция успешно завершена. Ждём вас снова.")
}

func filenamesFromCommandLine() (inFilename, outFilename string, err error) {
	usage := "Пример: %s [<] [e_main.comi | m_main.comi]\nРезультатом компиляции является e_compiled.comi или m_compiled.comi."
	// Отладочные строки
	// fmt.Println("Аргументов:", len(os.Args), "\nАргументы:", os.Args)
	// fmt.Println("filepath.Base(os.Args[1]):", filepath.Base(os.Args[1]))
	// fmt.Println("filepath.Ext(os.Args[1]):", filepath.Ext(os.Args[1]))

	if len(os.Args) == 2 && strings.EqualFold(filepath.Ext(os.Args[1]), ".comi") {
		inFilename = filepath.Base(os.Args[1])
		fmt.Println("Входной файл:", inFilename)
		outFilename = string(inFilename[0]) + "_compiled.comi"
		fmt.Println("Выходной файл:", outFilename)
	} else if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		err = fmt.Errorf(usage, filepath.Base(os.Args[0]))
		return "", "", err
	} else {
		fmt.Printf("Программа %s принимает не более одного аргумента.\n", filepath.Base(os.Args[0]))
		err = fmt.Errorf(usage, filepath.Base(os.Args[0]))
		return "", "", err
	}

	if inFilename != "" && inFilename == outFilename {
		log.Fatal("won’t overwrite the infile")
	}

	return inFilename, outFilename, nil

}

func compile(inFile io.Reader, outFile io.Writer) (err error) {

	reader := bufio.NewReader(inFile)
	writer := bufio.NewWriter(outFile)

	defer func() {
		if err == nil {
			err = writer.Flush()
		}
	}()

	eof := false // флаг конца файла в основном скрипте
	for !eof {
		/* dropline - это флаг "пропуска" прочитанной строки.
		Если dropline=false - прочитанная строка записывается в выходной файл,
		если dropline=true - читается следующая строка из входного файла*/
		dropline := false
		var line string
		line, err = reader.ReadString('\n')
		lineFields := strings.Fields(line)
		if len(lineFields) > 0 && (strings.EqualFold(lineFields[0], "$comi") || strings.EqualFold(lineFields[0], "$cominput")) {
			//-------------------------------------------
			// Вот здесь тоже надо сделать разбор путей
			//-------------------------------------------

			inScriptName := lineFields[1][1:len(lineFields[1])-1] + ".comi" // inline script name
			dropline = true

			// Здесь надо открыть файл для чтения с именем isname
			inScript := os.Stdin
			if inScript, err = os.Open(inScriptName); err != nil {
				log.Fatal(err)
			}
			defer inScript.Close()
			inReader := bufio.NewReader(inScript)

			// Здесь читаем вложенный скрипт по строкам
			ineof := false // флаг конца файла во вложенном скрипте
			for !ineof {
				var inline string
				inline, err = inReader.ReadString('\n')
				if err == io.EOF {
					err = nil    // в действительности признак io.EOF не является ошибкой
					ineof = true // это вызовет прекращение цикла в следующей итерации
				} else if err != nil {
					return err // в случае настоящей ошибки выйти немедленно
				}
				if _, err = writer.WriteString(inline); err != nil {
					return err
				}
			}
		}
		if err == io.EOF {
			err = nil  // в действительности признак io.EOF не является ошибкой
			eof = true // это вызовет прекращение цикла в следующей итерации
		} else if err != nil {
			return err // в случае настоящей ошибки выйти немедленно
		}

		if dropline == true {
			line, err = reader.ReadString('\n')
		}

		if _, err = writer.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}
