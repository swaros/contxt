package runner

func NewDefaultExplainer() *ExplainLib {
	lib := NewExplainLib()

	// Explaining: template: contxt-functions:4: function "include" not defined
	lib.AddExplainer(ErrExplainer{
		RefLine: "template: contxt-functions:[LN]: ",
		Lookup:  LineNumberIndicator,
		Explain: func(errParser *ErrParse, errRef ErrorReference) (string, bool) {
			// we have a template error
			// fist try to parse something like this: template: contxt-functions:4: function "include" not defined
			return extractCodeHelper(errParser.session.TemplateHndl.GetAviableSource(), errParser, errRef)
		},
		Info: `Verify the code of the template file. this type of error can happen if we try to parse templates
		they have a different use case then the default go templates. as example helm templates.
		You can still use this type of templates but if this contains unsupported markups, you have to
		create a "tpl.ignore" file in the same directory as the template file.
		put any unsupported markup in this file and the parser will ignore them.`,
	})

	// Explaining the error: yaml: line [LN]: did not find expected '-' indicator
	lib.AddExplainer(ErrExplainer{
		RefLine: YamlLineIndicator,
		Lookup:  LineNumberIndicator,
		Explain: func(errParser *ErrParse, errRef ErrorReference) (string, bool) {
			return extractCodeHelper(errParser.session.TemplateHndl.GetAviableSource(), errParser, errRef)
		},
		Info: `Verify the yaml file. if you have a yaml file that is not correctly formatted, you will get this error.
		You can use online tools like https://yamlchecker.com/ to verify the yaml file.
		the YAML file could also be invalid after the template parsing. if you have a template file that
		contains yaml code, you have to verify the output of the template parser.`,
	})
	return lib
}
