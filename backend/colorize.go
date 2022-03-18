package main

// imports {{{
import (
	"fmt"
)

//}}}

// variables {{{
var (
	Creset  string = "\033[0m"
	Cred    string = "\033[31m"
	Cgreen  string = "\033[32m"
	Cyellow string = "\033[33m"
	Cblue   string = "\033[34m"
	Cpurple string = "\033[35m"
	Ccyan   string = "\033[36m"
	Cwhite  string = "\033[37m"
)

//}}}

// string to colors {{{
func green(input_string string) string {
	return Cgreen + input_string + Creset
}

func red(input_string string) string {
	return Cred + input_string + Creset
}

func yellow(input_string string) string {
	return Cyellow + input_string + Creset
}

func blue(input_string string) string {
	return Cblue + input_string + Creset
}

func cyan(input_string string) string {
	return Ccyan + input_string + Creset
}

func purple(input_string string) string {
	return Cpurple + input_string + Creset
}

func white(input_string string) string {
	return Cwhite + input_string + Creset
}

//}}}

// debug prints {{{
func Debugf(s string, args ...interface{}) {
	fmt.Printf(cyan(" 💧 Debug: "+s+"\n"), args...)
}

func Debugln(s interface{}) {
	fmt.Print(cyan(" 💧 Debug: " + fmt.Sprintln(s)))
}

//}}}

// warning prints {{{
func Warningf(s string, args ...interface{}) {
	fmt.Printf(yellow(" ⚠️  Warning: "+s+"\n"), args...)
}

func Warningln(s interface{}) {
	fmt.Print(yellow(" ⚠️  Warning: " + fmt.Sprintln(s)))
}

//}}}

// error prints {{{
func Errorf(s string, args ...interface{}) {
	fmt.Printf(red(" ❌  Error: "+s+"\n"), args...)
}

func Errorln(s interface{}) {
	fmt.Print(red(" ❌  Error: " + fmt.Sprintln(s)))
}

//}}}

// success prints {{{
func Successf(s string, args ...interface{}) {
	fmt.Printf(green(" ✅ Success: "+s+"\n"), args...)
}

func Successln(s interface{}) {
	fmt.Print(green(" ✅ Success: " + fmt.Sprintln(s)))
}

//}}}

//+ test function
// func testColors() {

// 	Debugf("testing debug + %s", "stringa maggica")

// 	Warningf("testing debug + %s", "stringa maggica")

// 	Errorf("testing debug + %s", "stringa maggica")

// 	Successf("testing debug + %s", "stringa maggica")
// }
