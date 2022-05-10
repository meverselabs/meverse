package util

import (
	"log"

	"github.com/onsi/ginkgo/v2"
)

func Println(v ...interface{}) {
	log.Println(v...)
}

func PrintlnT(title string, v ...interface{}) {
	log.Print(title + " : ")
	log.Print(v...)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func GPrintln(v ...interface{}) {
	ginkgo.GinkgoWriter.Println(v...)
}

func GPrintlnT(title string, v ...interface{}) {
	ginkgo.GinkgoWriter.Print(title + " : ")
	ginkgo.GinkgoWriter.Println(v...)
}

func GPrintf(format string, v ...interface{}) {
	ginkgo.GinkgoWriter.Printf(format, v...)
}
