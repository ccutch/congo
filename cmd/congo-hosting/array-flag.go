package main

import "fmt"

type stringArray []string

func (i *stringArray) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *stringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}
