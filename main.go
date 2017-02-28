package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

func main() {
	var (
		indexFile  = flag.String("index", "", "index file (required)")
		configFile = flag.String("config", "", "config file (required)")
		toDir      = flag.String("to", "", "target dir (required)")
	)
	flag.Parse()

	if *indexFile == "" || *configFile == "" || *toDir == "" {
		fmt.Fprintf(os.Stderr, "Error: reqired parameter not specified\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	index, err := readIndex(*indexFile)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		os.Exit(1)
	}
	indexDir := filepath.Dir(*indexFile)

	config, err := readConfig(*configFile)
	if err != nil {
		log.Printf("Error reading data: %v", err)
		os.Exit(1)
	}

	err = generateFiles(index, config, indexDir, *toDir)
	if err != nil {
		log.Printf("Error generating files: %v", err)
		os.Exit(1)
	}
}

func generateFiles(index index, config interface{}, indexDir, toDir string) error {
	//Apply config to targets filepath
	for i := range index.Files {
		t, err := tmplToString(index.Files[i].Target, config)
		if err != nil {
			return fmt.Errorf("Error in file mapping: %v", err)
		}
		index.Files[i].Target = filepath.Join(toDir, t)
	}

	for _, mapping := range index.Files {
		fmt.Printf("%#v\n", mapping)
		switch {
		case mapping.Before != "":
			//insert snippet into file
			renderedSnippet, err := tmplFileToString(mapping.Template, config)
			if err != nil {
				return fmt.Errorf("Error in file mapping: %v", err)
			}
			err = insertBefore(mapping.Target, mapping.Before, renderedSnippet)
			if err != nil {
				return fmt.Errorf("Error inserting into %q:  %v", mapping.Target, err)
			}
		default:
			//create new file
			err := tmplFileToFile(filepath.Join(indexDir, mapping.Template), mapping.Target, config)
			if err != nil {
				return fmt.Errorf("Error applying template: %v", err)
			}
		}
	}
	return nil
}

func insertBefore(target, pattern string, snippet string) error {
	ptn, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	f, err := os.Open(target)
	if err != nil {
		return fmt.Errorf("could not open taget file %s: %v", target, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !ptn.Match(line) {
			buf.Write(line)
			buf.WriteByte('\n')
			continue
		}
		buf.WriteString(snippet)
		buf.WriteByte('\n')
		buf.Write(line)
		buf.WriteByte('\n')
		for scanner.Scan() {
			buf.Write(scanner.Bytes())
			buf.WriteByte('\n')
		}
		f.Close()
		ioutil.WriteFile(target, buf.Bytes(), 0644)
		break
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error during inserting snippet into %q:%v", target, err)
	}
	return nil
}

func tmplFileToString(tmplFile string, data interface{}) (string, error) {
	var buf bytes.Buffer
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %v", tmplFile, err)
	}
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("executing template %s: %v", tmplFile, err)
	}
	return buf.String(), nil
}

func tmplFileToFile(tmplFile, target string, data interface{}) error {
	targetDir := filepath.Dir(target)
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("creating target dir %q:%v", targetDir, err)
	}
	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("creating target file %q:%v", target, err)
	}
	defer f.Close()
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		return fmt.Errorf("parsing template %s: %v", tmplFile, err)
	}
	return tmpl.Execute(f, data)
}

func tmplToString(textTemplate string, data interface{}) (string, error) {
	tmpl := template.New("")
	t, err := tmpl.Parse(textTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing template %q: %v", textTemplate, err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("executing template %q with data %v: %v", textTemplate, data, err)
	}
	return buf.String(), nil
}

func readConfig(file string) (interface{}, error) {
	f, err := os.Open(file)
	if err != nil {
		return index{}, fmt.Errorf("opening data file %s: %v", file, err)
	}
	defer f.Close()
	var v interface{}
	err = json.NewDecoder(f).Decode(&v)
	if err != nil {
		return index{}, fmt.Errorf("decoding data file %s: %v", file, err)
	}
	return v, nil
}

func readIndex(file string) (index, error) {
	fmt.Printf("config %#v\n", file)
	f, err := os.Open(file)
	if err != nil {
		return index{}, fmt.Errorf("opening config file %s: %v", file, err)
	}
	defer f.Close()
	var v index
	err = json.NewDecoder(f).Decode(&v)
	if err != nil {
		return index{}, fmt.Errorf("decoding config file %s: %v", file, err)
	}
	return v, nil
}

type index struct {
	Files []fileMapping
}

type fileMapping struct {
	Template string `json:"from"`
	Target   string `json:"to"`
	Before   string `json:"before"`
}
