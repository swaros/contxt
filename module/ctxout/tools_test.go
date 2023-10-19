package ctxout_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestClearString(t *testing.T) {
	str := ctxout.StringCleanEscapeCodes("this is a test")
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}

	str = ctxout.StringCleanEscapeCodes("this is a \x1b[31mtest\x1b[0m")
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}
	colCode := ctxout.ToString(ctxout.NewMOWrap(), "this is a ", ctxout.BackBlack, "test\n", ctxout.CleanTag)
	str = ctxout.StringPure(colCode)
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}

	str = ctxout.StringCleanEscapeCodes(ctxout.ToString(ctxout.NewMOWrap(), "this is a ", ctxout.BackBlack, "test", ctxout.CleanTag))
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}
}

func TestStringLengthPrintable(t *testing.T) {

	type lenTest struct {
		in  string
		out int
	}

	tests := []lenTest{
		{"this is a test", 14},                // rnd 0
		{"this is a \x1b[31mtest\x1b[0m", 14}, // rnd 1
		{"你好世界", 8},                           // rnd 2
		{"🖵", 1},                              // rnd 3
		{"🖵\x1b[31m", 1},                      // rnd 4
		{"🌎\x1b[31m🌎🖵", 5},                    // rnd 5
		{"\u2588", 1},                         // rnd 6
	}

	for rnd, test := range tests {
		strLen := ctxout.VisibleLen(test.in)
		if strLen != test.out {
			t.Errorf("[rnd %d] Expected %d but got %d [%s]", rnd, test.out, strLen, test.in)
		} else {
			t.Logf(" OK [rnd %d]", rnd)
		}

	}
}

func TestStringCut(t *testing.T) {

	testStr := "1234567890abcdefghijklmnopqrstuvwxyz"
	expexted := "1234567890"

	cutStr, rest := ctxout.StringCut(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "abcdefghijklmnopqrstuvwxyz" {
		t.Errorf("Expected '%s' but got '%s'", "abcdefghijklmnopqrstuvwxyz", rest)
	}

	testStr = "123456"
	expexted = "123456"

	cutStr, rest = ctxout.StringCut(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "" {
		t.Errorf("Expected '%s' but got '%s'", "", rest)
	}

}

func TestStringCutRight(t *testing.T) {

	testStr := "1234567890abcdefghijklmnopqrstuvwxyz"
	expexted := "qrstuvwxyz"

	cutStr, rest := ctxout.StringCutFromRight(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "1234567890abcdefghijklmnop" {
		t.Errorf("Expected '%s' but got '%s'", "1234567890abcdefghijklmnop", rest)
	}

	testStr = "123456"
	expexted = "123456"

	cutStr, rest = ctxout.StringCutFromRight(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "" {
		t.Errorf("Expected '%s' but got '%s'", "", rest)
	}

}

func TestFitWords(t *testing.T) {
	size := 80
	source := `ams-xml-create ams-xml-delete ams-xml-show-files ams-xml-update-java check-git-branch check-git-url check-have-changes check-permissions chown-all ci-build ci-protocol clean clean-rcon create-config create-db-model-files git-pull git-reset-created git-status git-submodule-update init laun-chunityhub local-dev maintainer print-unityhub-pid rcon rcon-start run-dev run-php-inside show-docker-compose start-unityhub stop-dev test unityhub-log unityhub-log-clean unityhub-log-tail verify write-docker-compose xml-check-crewmember-144 xml-check-crewmember-145 xml-check-crewmember-146 xml-check-crewmember-147 xml-check-crewmember-148 xml-check-crewmember-149 xml-check-crewmember-150 xml-check-crewmember-151 xml-check-crewmember-152 xml-check-crewmember-153 xml-check-crewmember-154 xml-check-crewmember-155 xml-check-extension-138 xml-check-extension-139 xml-check-extension-140 xml-check-extension-141 xml-check-extension-142 xml-check-extension-143 xml-check-extension-144 xml-check-extension-145 xml-check-extension-146 xml-check-extension-147 xml-check-extension-148 xml-check-extension-149 xml-check-extension-150 xml-check-extension-151 xml-check-extension-152 xml-check-extension-153 xml-check-extension-154 xml-check-extension-155 xml-check-extension-156 xml-check-extension-157 xml-check-extension-158 xml-check-extension-159 xml-check-extension-160 xml-check-extension-161 xml-check-extension-162 xml-check-extension-163 xml-check-extension-164 xml-check-extension-165 xml-check-extension-166 xml-check-extension-167 xml-check-extension-168 xml-check-extension-169 xml-check-extension-170 xml-check-extension-171 xml-check-extension-172 xml-check-extension-173 xml-check-extension-174 xml-check-extension-175 xml-check-extension-176 xml-check-extension-177 xml-check-extension-178 xml-display-crewmember-144 xml-display-crewmember-145 xml-display-crewmember-146 xml-display-crewmember-147 xml-display-crewmember-148 xml-display-crewmember-149 xml-display-crewmember-150 xml-display-crewmember-151 xml-display-crewmember-152 xml-display-crewmember-153 xml-display-crewmember-154 xml-display-crewmember-155 xml-display-extension-138 xml-display-extension-139 xml-display-extension-140 xml-display-extension-141 xml-display-extension-142 xml-display-extension-143 xml-display-extension-144 xml-display-extension-145 xml-display-extension-146 xml-display-extension-147 xml-display-extension-148 xml-display-extension-149 xml-display-extension-150 xml-display-extension-151 xml-display-extension-152 xml-display-extension-153 xml-display-extension-154 xml-display-extension-155 xml-display-extension-156 xml-display-extension-157 xml-display-extension-158 xml-display-extension-159 xml-display-extension-160 xml-display-extension-161 xml-display-extension-162 xml-display-extension-163 xml-display-extension-164 xml-display-extension-165 xml-display-extension-166 xml-display-extension-167 xml-display-extension-168 xml-display-extension-169 xml-display-extension-170 xml-display-extension-171 xml-display-extension-172 xml-display-extension-173 xml-display-extension-174 xml-display-extension-175 xml-display-extension-176 xml-display-extension-177 xml-display-extension-178 xml-remove-crewmember-144 xml-remove-crewmember-145 xml-remove-crewmember-146 xml-remove-crewmember-147 xml-remove-crewmember-148 xml-remove-crewmember-149 xml-remove-crewmember-150 xml-remove-crewmember-151 xml-remove-crewmember-152 xml-remove-crewmember-153 xml-remove-crewmember-154 xml-remove-crewmember-155 xml-remove-extension-138 xml-remove-extension-139 xml-remove-extension-140 xml-remove-extension-141 xml-remove-extension-142 xml-remove-extension-143 xml-remove-extension-144 xml-remove-extension-145 xml-remove-extension-146 xml-remove-extension-147 xml-remove-extension-148 xml-remove-extension-149 xml-remove-extension-150 xml-remove-extension-151 xml-remove-extension-152 xml-remove-extension-153 xml-remove-extension-154 xml-remove-extension-155 xml-remove-extension-156 xml-remove-extension-157 xml-remove-extension-158 xml-remove-extension-159 xml-remove-extension-160 xml-remove-extension-161 xml-remove-extension-162 xml-remove-extension-163 xml-remove-extension-164 xml-remove-extension-165 xml-remove-extension-166 xml-remove-extension-167 xml-remove-extension-168 xml-remove-extension-169 xml-remove-extension-170 xml-remove-extension-171 xml-remove-extension-172 xml-remove-extension-173 xml-remove-extension-174 xml-remove-extension-175 xml-remove-extension-176 xml-remove-extension-177 xml-remove-extension-178 xml-update-java-crewmember-144 xml-update-java-crewmember-145 xml-update-java-crewmember-146 xml-update-java-crewmember-147 xml-update-java-crewmember-148 xml-update-java-crewmember-149 xml-update-java-crewmember-150 xml-update-java-crewmember-151 xml-update-java-crewmember-152 xml-update-java-crewmember-153 xml-update-java-crewmember-154 xml-update-java-crewmember-155 xml-update-java-extension-138 xml-update-java-extension-139 xml-update-java-extension-140 xml-update-java-extension-141 xml-update-java-extension-142 xml-update-java-extension-143 xml-update-java-extension-144 xml-update-java-extension-145 xml-update-java-extension-146 xml-update-java-extension-147 xml-update-java-extension-148 xml-update-java-extension-149 xml-update-java-extension-150 xml-update-java-extension-151 xml-update-java-extension-152 xml-update-java-extension-153 xml-update-java-extension-154 xml-update-java-extension-155 xml-update-java-extension-156 xml-update-java-extension-157 xml-update-java-extension-158 xml-update-java-extension-159 xml-update-java-extension-160 xml-update-java-extension-161 xml-update-java-extension-162 xml-update-java-extension-163 xml-update-java-extension-164 xml-update-java-extension-165 xml-update-java-extension-166 xml-update-java-extension-167 xml-update-java-extension-168 xml-update-java-extension-169 xml-update-java-extension-170 xml-update-java-extension-171 xml-update-java-extension-172 xml-update-java-extension-173 xml-update-java-extension-174 xml-update-java-extension-175 xml-update-java-extension-176 xml-update-java-extension-177 xml-update-java-extension-178`
	newSource := ctxout.FitWordsToMaxLen(source, size)

	lines := strings.Split(newSource, "\n")
	for _, line := range lines {
		if len(line) > size {
			t.Errorf("Line '%s' is longer than %d chars (%d)", line, size, len(line))
		}
	}
	// test long word. the word should not be touched if it is longer than size
	// but we would get an newline after the word
	result := ctxout.FitWordsToMaxLen("ab 1234567890 cdef", 5)
	expected := "ab\n1234567890\ncdef"
	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}

func assertWordFits(t *testing.T, source string, size int) {
	newSource := ctxout.FitWordsToMaxLen(source, size)

	lines := strings.Split(newSource, "\n")
	for _, line := range lines {
		if len(line) > size {
			// line can be longer if it contains a word longer than size
			// so if the line is longer than size, it must contain a word longer than size
			words := strings.Split(line, " ")
			for _, word := range words {
				if len(word) > size {
					t.Logf("OK: Line '%s' is longer than %d chars (%d) because of word '%s'", line, size, len(line), word)
					return
				}
			}
			t.Errorf("Line '%s' is longer than %d chars (%d)", line, size, len(line))
		}
	}
}

func TestFitWords2(t *testing.T) {
	source := ` [SOT]nec consequat nec diam Vivamus consequat viverra.... viverra. Vivamus nisl consequat Vivamus viverra. diam [EOT]
	diam viverra. elit. nec dolor nisl         nec ipsum nec diam consequat elit. consequat viverra. consequat 
	consequat diam nec diam diam nisl viverra. Sed                 nisl elit
	. amet, nec amet, consequat diam Vivamus nec nisl Vivam
	viverra. nec nisl diam consectetur c
	onsequat diam nec nec      us Vivamus diam diam necnisl nisl nisl diam viverra. viverra. ip
	Vivamus nec diam viverra. viverra. viverra. nec consequat             sum diam viverra.diam viverra. nec Vivamus nec ipsum diam
	Vivamus    viverra. nisl
	nec elit. diam Vivamus viverra.diam diam nisl diam nisl nisl nisl Vivamus
	consequat consectetur viverra. nec nec nec elit. consequat     consequat nec do
	lor nisl Vivamus Vivamus Vivamus nisl nec conseq
	viverra. consequat                   uat viverra.Vivamus nec sit
	nec nec nec diam              viverra. nisl nec ipsum eu Vivamus
	nec Lorem Sed Vivamus  consequat diam nec Vivamus nec Vivamus amet, nisl nec nec conseq
	diam consequat diam diam amet, Vivamus diam consequat nec nec        uat n
	ec nec viverra. necdolor nec viverra.  the last [EOT]`
	assertWordFits(t, source, 10)
}

func TestFitWords3(t *testing.T) {
	source := `Checkifthissomething
	verylongword`
	assertWordFits(t, source, 10)
}
